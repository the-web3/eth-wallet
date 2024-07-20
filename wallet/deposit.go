package wallet

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
				continue
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
	blockListForStore := make([]database.Blocks, len(headers))
	var depositList []database.Deposits
	var withdrawList []database.Withdraws
	var depositTransactionList []database.Transactions
	var outherTransactionList []database.Transactions
	var tokenBalanceList []database.TokenBalance
	var batchLastBlockNumber uint64
	for i := range headers {
		log.Info("handle block number", "number", headers[i].Number.String(), "blockHash", headers[i].Hash().String())

		blockListForStore[i] = database.BlockHeaderFromHeader(&d.headers[i])

		block, err := d.client.BlockByNumber(headers[i].Number)
		if err != nil {
			log.Error("get block number error", "err", err)
			return err
		}
		deposits, withdraws, depositTransactions, outherTransactions, tokenBalances, err := d.processTransactions(block.Transactions, block.BaseFee)
		if err != nil {
			log.Error("process transaction fail", "err", err)
			return err
		}

		depositList = append(depositList, deposits...)
		withdrawList = append(withdrawList, withdraws...)
		depositTransactionList = append(depositTransactionList, depositTransactions...)
		outherTransactionList = append(outherTransactionList, outherTransactions...)
		tokenBalanceList = append(tokenBalanceList, tokenBalances...)
		batchLastBlockNumber = headers[i].Number.Uint64()
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](d.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := d.db.Transaction(func(tx *database.DB) error {
			if err := tx.Blocks.StoreBlockss(blockListForStore, uint64(len(blockListForStore))); err != nil {
				return err
			}

			if len(depositList) > 0 {
				log.Info("Store deposit transaction success", "totalTx", len(depositList))
				if err := tx.Deposits.StoreDeposits(depositList, uint64(len(depositList))); err != nil {
					return err
				}
			}
			log.Info("batch latest block number", "batchLastBlockNumber", batchLastBlockNumber)

			// 更新之前充值确认位
			if err := tx.Deposits.UpdateDepositsStatus(batchLastBlockNumber - uint64(d.chainConf.Confirmations)); err != nil {
				return err
			}

			if len(withdrawList) > 0 {
				if err := tx.Withdraws.UpdateTransactionStatus(withdrawList); err != nil {
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

			if len(tokenBalanceList) > 0 {
				log.Info("update or store token balance", "tokenBalanceList", len(tokenBalanceList))
				if err := tx.Balances.UpdateOrCreate(tokenBalanceList); err != nil {
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

func (d *Deposit) processTransactions(txList []node.TransactionList, baseFee string) ([]database.Deposits, []database.Withdraws, []database.Transactions, []database.Transactions, []database.TokenBalance, error) {
	if len(txList) == 0 {
		log.Error("no transactions")
		return nil, nil, nil, nil, nil, errors.New("no transactions")
	}
	var depositList []database.Deposits
	var withdrawList []database.Withdraws
	var depositTransactionList []database.Transactions
	var otherTransactionList []database.Transactions
	var tokenBalanceList []database.TokenBalance
	for _, tx := range txList {
		txHash := tx.Hash
		var isToken bool
		tokens, err := d.db.Tokens.TokensInfoByAddress(tx.To)
		if err != nil {
			log.Error("query token info fail", "err", err)
			continue
		}
		log.Info("=query token by address success=", "tokens", tokens)
		txTo := common.HexToAddress(tx.To)
		addrTo, err := d.db.Addresses.QueryAddressesByToAddress(&txTo)
		if err != nil {
			log.Error("query address info fail", "err", err)
			continue
		}
		if tokens == nil && addrTo == nil {
			continue
		}

		transaction, err := d.client.TxByHash(common.HexToHash(txHash))
		if err != nil {
			log.Error("get tx fail", err)
			return nil, nil, nil, nil, nil, err
		}
		signer := types.LatestSignerForChainID(big.NewInt(int64(d.chainConf.ChainID)))
		if err != nil {
			log.Error("Failed to get transaction message: %v", err)
		}
		fromAddress, _ := signer.Sender(transaction)

		// 当 to 地址为空时候，这种交易创建合约交易
		if transaction.To() == nil {
			log.Info("this is a contract create")
			continue
		}

		var toAddress common.Address
		var tokenAddress common.Address
		var decValue *big.Int
		// 当 to 地址为 token 地址的时候，这种交易为 token 充值，如果 to 地址不是合约地址就是 ETH 充值

		if tokens != nil {
			isToken = true
			inputData := hexutil.Encode(transaction.Data()[:])
			if len(inputData) < 138 {
				continue
			}
			if inputData[:10] != "0xa9059cbb" {
				continue
			}
			toAddress = common.HexToAddress("0x" + inputData[34:74])

			trimHex := strings.TrimLeft(inputData[74:138], "0")
			if len(trimHex) <= 0 {
				continue
			}

			rawValue, err := hexutil.DecodeBig("0x" + trimHex)
			if err != nil {
				log.Error("decode big int fail", "err", err)
				continue
			}
			tokenAddress = *transaction.To()
			decValue = decimal.NewFromBigInt(rawValue, 0).BigInt()
		} else {
			toAddress = *transaction.To()
			tokenAddress = common.Address{}
			decValue = transaction.Value()
		}

		addressTo, err := d.db.Addresses.QueryAddressesByToAddress(&toAddress)
		if err != nil {
			log.Error("query to address from addresses table fail", "err", err)
		}
		addressFrom, err := d.db.Addresses.QueryAddressesByToAddress(&fromAddress)
		if err != nil {
			log.Error("query from address from addresses table fail", "err", err)
		}
		if addressTo == nil && addressFrom == nil {
			log.Info("no transaction relate to wallet")
			continue
		}

		txReceipt, err := d.client.TxReceiptByHash(common.HexToHash(txHash))
		if err != nil {
			log.Error("get tx fail", err)
			return nil, nil, nil, nil, nil, err
		}
		log.Info("============================================================")
		log.Info("handle transaction success", "txHash", transaction.Hash().String(), "txReceiptHash", txReceipt.TxHash.String())
		log.Info("============================================================")

		// withdraw
		withdraw, err := d.db.Withdraws.QueryWithdrawsByHash(transaction.Hash())
		if err != nil {
			log.Error("query withdraw transaction fail", "err", err)
			continue
		}

		// collection cold
		ccTx, err := d.db.Transactions.QueryTransactionByHash(transaction.Hash())
		if err != nil {
			log.Error("query withdraw transaction fail", "err", err)
			continue
		}
		var gasPrice *big.Int
		var transactionFee *big.Int
		if (addressTo != nil && txReceipt.Status == 1) || (ccTx != nil && txReceipt.Status == 1) || (withdraw != nil && txReceipt.Status == 1) {
			if txReceipt.Type == types.DynamicFeeTxType {
				gasPrice = txReceipt.EffectiveGasPrice
				baseFeeInt, _ := strconv.ParseInt(baseFee, 10, 64)
				transactionFee = new(big.Int).Add(gasPrice, big.NewInt(baseFeeInt))
				transactionFee.Mul(transactionFee, new(big.Int).SetUint64(txReceipt.GasUsed))
			} else {
				gasPrice = txReceipt.EffectiveGasPrice
				transactionFee = new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(txReceipt.GasUsed))
			}

			// 充值：to 是系统用户地址， from 地址是外部地址
			if addressTo != nil && txReceipt.Status == 1 && addressFrom == nil {
				log.Info("Find Deposit transaction", "TxHash", transaction.Hash().String())
				deposit, err := d.HandleDeposit(transaction, txReceipt, transactionFee, isToken, decValue, fromAddress, toAddress, tokenAddress)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, nil, err
				}
				depositList = append(depositList, deposit)
				tx, tokenBalance, err := d.HandleTransaction(transaction, txReceipt, transactionFee, 0, isToken, decValue, fromAddress, toAddress, tokenAddress)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, nil, err
				}
				depositTransactionList = append(depositTransactionList, tx)
				tokenBalanceList = append(tokenBalanceList, tokenBalance)
			}

			// 提现：from 地址系统的热钱包地址，to 地址是外部地址
			if withdraw != nil && txReceipt.Status == 1 && addressFrom != nil && addressTo == nil {
				log.Info("Find withdraw transaction", "TxHash", transaction.Hash().String())
				withdrawItem, err := d.HandleWithdaw(transaction, txReceipt, transactionFee, isToken, decValue, fromAddress, toAddress, tokenAddress)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, nil, err
				}
				withdrawList = append(withdrawList, withdrawItem)
				tx, tokenBalance, err := d.HandleTransaction(transaction, txReceipt, transactionFee, 1, isToken, decValue, fromAddress, toAddress, tokenAddress)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, nil, err
				}
				otherTransactionList = append(otherTransactionList, tx)
				tokenBalanceList = append(tokenBalanceList, tokenBalance)
			}

			// 归集：to 地址是系统热钱包地址， from 地址系统用户
			// 热转冷：from 是系统的热钱包地址，to 地址是系统的冷钱包地址
			if ccTx != nil && txReceipt.Status == 1 && addressFrom != nil && addressTo != nil {
				tx, tokenBalance, err := d.HandleTransaction(transaction, txReceipt, transactionFee, 2, isToken, decValue, fromAddress, toAddress, tokenAddress)
				if err != nil {
					log.Error("handle deposit error", "err", err)
					return nil, nil, nil, nil, nil, err
				}
				otherTransactionList = append(otherTransactionList, tx)
				tokenBalanceList = append(tokenBalanceList, tokenBalance)
			}
		}
	}
	return depositList, withdrawList, depositTransactionList, otherTransactionList, tokenBalanceList, nil
}

func (d *Deposit) HandleDeposit(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int, isToken bool, decValue *big.Int, fromAddr, toAddr, tokenAddress common.Address) (database.Deposits, error) {
	if transaction == nil || receipt == nil {
		return database.Deposits{}, errors.New("transation or receipt is empty")
	}
	var amount *big.Int
	var toAddress common.Address
	if !isToken {
		amount = transaction.Value()
		toAddress = *transaction.To()
	} else {
		amount = decValue
		toAddress = toAddr
	}
	guid, _ := uuid.NewUUID()
	deposit := database.Deposits{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      fromAddr,
		ToAddress:        toAddress,
		TokenAddress:     tokenAddress,
		Fee:              Fee,
		Amount:           amount,
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		Timestamp:        uint64(transaction.Time().Unix()),
	}
	return deposit, nil
}

func (d *Deposit) HandleWithdaw(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int, isToken bool, decValue *big.Int, fromAddr, toAddr, tokenAddress common.Address) (database.Withdraws, error) {
	if transaction == nil || receipt == nil {
		return database.Withdraws{}, errors.New("transation or receipt is empty")
	}
	var amount *big.Int
	var toAddress common.Address
	if !isToken {
		amount = transaction.Value()
		toAddress = *transaction.To()
	} else {
		amount = decValue
		toAddress = toAddr
	}
	guid, _ := uuid.NewUUID()
	withdraw := database.Withdraws{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      fromAddr,
		ToAddress:        toAddress,
		TokenAddress:     tokenAddress,
		Fee:              Fee,
		Amount:           amount,
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		TxSignHex:        "rsv",
		Timestamp:        uint64(transaction.Time().Unix()),
	}
	return withdraw, nil
}

func (d *Deposit) HandleTransaction(transaction *types.Transaction, receipt *types.Receipt, Fee *big.Int, txtype uint8, isToken bool, decValue *big.Int, fromAddr, toAddr, tokenAddress common.Address) (database.Transactions, database.TokenBalance, error) {
	if transaction == nil || receipt == nil {
		return database.Transactions{}, database.TokenBalance{}, errors.New("transation or receipt is empty")
	}
	var amount *big.Int
	var toAddress common.Address
	if !isToken {
		amount = transaction.Value()
		toAddress = *transaction.To()
	} else {
		amount = decValue
		toAddress = toAddr
	}
	guid, _ := uuid.NewUUID()
	tx := database.Transactions{
		GUID:             guid,
		BlockHash:        receipt.BlockHash,
		BlockNumber:      receipt.BlockNumber,
		Hash:             transaction.Hash(),
		FromAddress:      fromAddr,
		ToAddress:        toAddress,
		TokenAddress:     toAddress,
		Fee:              Fee,
		Amount:           amount,
		Status:           uint8(receipt.Status),
		TransactionIndex: big.NewInt(int64(receipt.TransactionIndex)),
		TxType:           txtype,
		Timestamp:        uint64(transaction.Time().Unix()),
	}

	balance := database.TokenBalance{
		Address:      toAddr,
		TokenAddress: tokenAddress,
		Balance:      transaction.Value(),
		LockBalance:  big.NewInt(0),
		TxType:       txtype,
	}
	return tx, balance, nil
}
