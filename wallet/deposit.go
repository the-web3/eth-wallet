package wallet

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/the-web3/eth-wallet/common/tasks"
	"github.com/the-web3/eth-wallet/config"
	"github.com/the-web3/eth-wallet/database"
	"github.com/the-web3/eth-wallet/wallet/node"
	"github.com/the-web3/eth-wallet/wallet/retry"
)

type Deposit struct {
	db        *database.DB
	chainConf *config.ChainConfig

	client          node.EthClient
	headerTraversal *node.HeaderTraversal

	headers []types.Header

	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
}

func NewDeposit(cfg *config.Config, db *database.DB, client node.EthClient, shutdown context.CancelCauseFunc) (*Deposit, error) {
	latestHeader, err := db.Blocks.LatestBlocks()
	if err != nil {
		return nil, err
	}
	var fromHeader *types.Header
	if latestHeader != nil {
		log.Info("sync detected last indexed block", "number", latestHeader.Number, "hash", latestHeader.Hash)
		fromHeader = latestHeader.RLPHeader.Header()
	} else if cfg.Chain.BlocksStep > 0 {
		log.Info("no sync indexed state starting from supplied ethereum height", "height", cfg.Chain.StartingHeight)
		header, err := client.BlockHeaderByNumber(big.NewInt(int64(cfg.Chain.StartingHeight)))
		if err != nil {
			return nil, fmt.Errorf("could not fetch starting block header: %w", err)
		}
		fromHeader = header
	} else {
		log.Info("no eth wallet indexed state")
	}
	headerTraversal := node.NewHeaderTraversal(client, fromHeader, big.NewInt(int64(cfg.Chain.BlocksStep)), cfg.Chain.ChainID)

	resCtx, resCancel := context.WithCancel(context.Background())

	return &Deposit{
		db:              db,
		chainConf:       &cfg.Chain,
		client:          client,
		headerTraversal: headerTraversal,
		resourceCtx:     resCtx,
		resourceCancel:  resCancel,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in deposit: %w", err))
		}},
	}, nil
}

func (d *Deposit) Close() error {
	var result error
	d.resourceCancel()
	if err := d.tasks.Wait(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to await deposit %w"), err)
		return result
	}
	return nil
}

func (d *Deposit) Start() error {
	log.Info("start deposit......")
	tickerDepositWorker := time.NewTicker(time.Second * 5)
	d.tasks.Go(func() error {
		for range tickerDepositWorker.C {
			newHeaders, err := d.headerTraversal.NextHeaders(uint64(d.chainConf.BlocksStep))
			if err != nil {
				log.Error("error querying for headers", "err", err)
			} else if len(newHeaders) == 0 {
				log.Warn("no new headers. syncer at head?")
			} else {
				d.headers = newHeaders
			}
			err = d.processBatch(d.headers)
			if err == nil {
				d.headers = nil
			}
		}
		return nil
	})
	return nil
}

func (d *Deposit) processBatch(headers []types.Header) error {
	var blockListForStore []database.Blocks
	var depositList []database.Deposit
	var withdrawList []database.Withdraw
	var depositTransactionList []database.Transactions
	var outherTransactionList []database.Transactions
	var batchLastBlockNumber uint64
	for i := range headers {
		log.Info("handle block number", "number", headers[i].Number.String(), "blockHash", headers[i].Hash().String())
		blockListForStore[i] = database.BlockHeaderFromHeader(&d.headers[i])

		block, err := d.client.BlockByNumber(headers[i].Number)
		if err != nil {
			log.Error("get block number error", "err", err)
			return err
		}
		depositList, withdrawList, depositTransactionList, outherTransactionList, err = d.processTransactions(block.Transactions, block.BaseFee)
		if err != nil {
			log.Error("process transaction fail", "err", err)
			return err
		}
		batchLastBlockNumber = headers[i].Number.Uint64()
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](d.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := d.db.Transaction(func(tx *database.DB) error {
			if err := tx.Blocks.StoreBlockss(blockListForStore, uint64(len(blockListForStore))); err != nil {
				return err
			}

			if len(depositList) > 0 {
				if err := tx.Deposit.StoreDeposits(depositList, uint64(len(depositList))); err != nil {
					return err
				}
			}
			// 更新之前充值确认位
			if err := tx.Deposit.UpdateDepositStatus(batchLastBlockNumber - uint64(d.chainConf.Confirmations)); err != nil {
				return err
			}

			if len(withdrawList) > 0 {
				if err := tx.Withdraw.UpdateTransactionStatus(withdrawList); err != nil {
					return err
				}
			}

			if len(depositTransactionList) > 0 {
				if err := tx.Transactions.StoreTransactions(depositTransactionList, uint64(len(depositTransactionList))); err != nil {
					return err
				}
			}

			if len(outherTransactionList) > 0 { // 提现和归集
				if err := tx.Transactions.UpdateTransactionStatus(outherTransactionList); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			log.Error("unable to persist batch", "err", err)
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return err
	}
	return nil
}

func (d *Deposit) processTransactions(txList []string, baseFee string) ([]database.Deposit, []database.Withdraw, []database.Transactions, []database.Transactions, error) {
	if len(txList) == 0 {
		log.Error("no transactions")
		return nil, nil, nil, nil, errors.New("no transactions")
	}
	var depositList []database.Deposit
	var withdrawList []database.Withdraw
	var depositTransactionList []database.Transactions
	var otherTransactionList []database.Transactions
	for _, txHash := range txList {
		transaction, err := d.client.TxByHash(common.HexToHash(txHash))
		if err != nil {
			log.Error("get tx fail", err)
			return nil, nil, nil, nil, err
		}

		txReceipt, err := d.client.TxReceiptByHash(common.HexToHash(txHash))
		if err != nil {
			log.Error("get tx fail", err)
			return nil, nil, nil, nil, err
		}
		log.Info("handle transaction success", "txHash", transaction.Hash().String(), "txReceiptHash", txReceipt.TxHash.String())

		// deposit
		address, err := d.db.Addresses.QueryAddressesByToAddres(transaction.To())
		if err != nil {
			log.Error("query address from addresses table fail", "err", err)
			return nil, nil, nil, nil, err
		}

		// withdraw
		withdraw, err := d.db.Withdraw.QueryWithdrawByHash(transaction.Hash())
		if err != nil {
			log.Error("query withdraw transaction fail", "err", err)
			return nil, nil, nil, nil, err
		}

		// collection cold
		ccTx, err := d.db.Transactions.QueryTransactionByHash(transaction.Hash())
		if err != nil {
			log.Error("query withdraw transaction fail", "err", err)
			return nil, nil, nil, nil, err
		}
		var gasPrice *big.Int
		var transactionFee *big.Int
		if (address != nil && txReceipt.Status == 1) || (ccTx != nil && txReceipt.Status == 1) || (withdraw != nil && txReceipt.Status == 1) {
			var txType uint8
			if txReceipt.Type == types.DynamicFeeTxType {
				gasPrice = txReceipt.EffectiveGasPrice
				baseFeeInt, _ := strconv.ParseInt(baseFee, 10, 64)
				transactionFee = new(big.Int).Add(gasPrice, big.NewInt(baseFeeInt))
				transactionFee.Mul(transactionFee, new(big.Int).SetUint64(txReceipt.GasUsed))
			} else {
				gasPrice = txReceipt.EffectiveGasPrice
				transactionFee = new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(txReceipt.GasUsed))
			}
			if address != nil && txReceipt.Status == 1 {
				deposit, err := d.HandleDeposit(transaction, txReceipt, transactionFee)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, err
				}
				depositList = append(depositList, deposit)
				tx, err := d.HandleTransaction(transaction, txReceipt, transactionFee, txType)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, err
				}
				depositTransactionList = append(depositTransactionList, tx)
			}

			if withdraw != nil && txReceipt.Status == 1 {
				withdrawItem, err := d.HandleWithdaw(transaction, txReceipt, transactionFee)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, err
				}
				withdrawList = append(withdrawList, withdrawItem)
				tx, err := d.HandleTransaction(transaction, txReceipt, transactionFee, 1)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, err
				}
				otherTransactionList = append(otherTransactionList, tx)
			}
			if ccTx != nil && txReceipt.Status == 1 {
				tx, err := d.HandleTransaction(transaction, txReceipt, transactionFee, 2)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, err
				}
				otherTransactionList = append(otherTransactionList, tx)
			}
		}
	}
	return depositList, withdrawList, depositTransactionList, otherTransactionList, nil
}

func (d *Deposit) HandleDeposit(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int) (database.Deposit, error) {
	if transaction == nil || receipt == nil {
		return database.Deposit{}, errors.New("transation or receipt is empty")
	}
	guid, _ := uuid.NewUUID()
	return database.Deposit{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      *transaction.To(),
		ToAddress:        *transaction.To(),
		Fee:              Fee.String(),
		Amount:           transaction.Value().String(),
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		R:                "r",
		S:                "s",
		V:                "v",
		Timestamp:        100000,
	}, nil
}

func (d *Deposit) HandleWithdaw(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int) (database.Withdraw, error) {
	if transaction == nil || receipt == nil {
		return database.Withdraw{}, errors.New("transation or receipt is empty")
	}
	guid, _ := uuid.NewUUID()
	return database.Withdraw{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      *transaction.To(),
		ToAddress:        *transaction.To(),
		Fee:              Fee.String(),
		Amount:           transaction.Value().String(),
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		TxSignHex:        "rsv",
		R:                "r",
		S:                "s",
		V:                "v",
		Timestamp:        100000,
	}, nil
}

func (d *Deposit) HandleTransaction(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int, txtype uint8) (database.Transactions, error) {
	if transaction == nil || receipt == nil {
		return database.Transactions{}, errors.New("transation or receipt is empty")
	}
	guid, _ := uuid.NewUUID()
	return database.Transactions{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      *transaction.To(),
		ToAddress:        *transaction.To(),
		Fee:              Fee.String(),
		Amount:           transaction.Value().String(),
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		TxType:           txtype,
		R:                "r",
		S:                "s",
		V:                "v",
		Timestamp:        100000,
	}, nil
}
