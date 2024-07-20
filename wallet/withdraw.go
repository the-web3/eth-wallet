package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/the-web3/eth-wallet/wallet/retry"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/the-web3/eth-wallet/wallet/ethereum"

	"github.com/the-web3/eth-wallet/common/tasks"
	"github.com/the-web3/eth-wallet/config"
	"github.com/the-web3/eth-wallet/database"
	"github.com/the-web3/eth-wallet/wallet/node"
)

var (
	EthGasLimit          uint64 = 21000
	TokenGasLimit        uint64 = 120000
	maxFeePerGas                = big.NewInt(2900000000)
	maxPriorityFeePerGas        = big.NewInt(2600000000)
)

type Withdraw struct {
	db             *database.DB
	chainConf      *config.ChainConfig
	client         node.EthClient
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
}

func NewWithdraw(cfg *config.Config, db *database.DB, client node.EthClient, shutdown context.CancelCauseFunc) (*Withdraw, error) {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &Withdraw{
		db:             db,
		chainConf:      &cfg.Chain,
		client:         client,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in deposit: %w", err))
		}},
	}, nil
}

func (w *Withdraw) Close() error {
	var result error
	w.resourceCancel()
	if err := w.tasks.Wait(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to await deposit %w"), err)
	}
	return nil
}

func (w *Withdraw) Start() error {
	log.Info("start withdraw......")
	tickerWithdrawsWorker := time.NewTicker(time.Second * 5)
	w.tasks.Go(func() error {
		for range tickerWithdrawsWorker.C {
			withdrawList, err := w.db.Withdraws.UnSendWithdrawsList()
			if err != nil {
				log.Error("get unsend withdraw list fail", "err", err)
				return err
			}

			returnWithdrawsList := make([]database.Withdraws, len(withdrawList))
			index := 0
			var balanceList []database.Balances
			for _, withdraw := range withdrawList {
				hotWallet, err := w.db.Addresses.QueryHotWalletInfo()
				if err != nil {
					log.Error("query hot wallet info err", "err", err)
					return err
				}

				hotWalletTokenBalance, err := w.db.Balances.QueryWalletBalanceByTokenAndAddress(hotWallet.Address, withdraw.TokenAddress)
				if hotWalletTokenBalance.Balance.Cmp(withdraw.Amount) < 0 {
					log.Info("hot wallet balance is not enough", "tokenAddress", withdraw.TokenAddress)
					continue
				}

				nonce, err := w.client.TxCountByAddress(hotWallet.Address)
				if err != nil {
					log.Error("query nonce by address fail", "err", err)
					return err
				}

				var buildData []byte
				var gasLimit uint64
				var toAddress *common.Address
				var amount *big.Int
				if withdraw.TokenAddress.Hex() != "0x0000000000000000000000000000000000000000" {
					buildData = ethereum.BuildErc20Data(withdraw.ToAddress, withdraw.Amount)
					toAddress = &withdraw.TokenAddress
					gasLimit = TokenGasLimit
					amount = big.NewInt(0)
				} else {
					toAddress = toAddress
					gasLimit = EthGasLimit
					amount = withdraw.Amount
				}
				dFeeTx := &types.DynamicFeeTx{
					ChainID:   big.NewInt(int64(w.chainConf.ChainID)),
					Nonce:     uint64(nonce),
					GasTipCap: maxPriorityFeePerGas,
					GasFeeCap: maxFeePerGas,
					Gas:       gasLimit,
					To:        toAddress,
					Value:     amount,
					Data:      buildData,
				}
				rawTx, txHash, err := ethereum.OfflineSignTx(dFeeTx, hotWallet.PrivateKey, big.NewInt(int64(w.chainConf.ChainID)))
				if err != nil {
					log.Error("offline transaction fail", "err", err)
					return err
				}
				log.Info("Offline sign tx success", "rawTx", rawTx)

				// sendRawTx
				err = w.client.SendRawTransaction(rawTx)
				if err != nil {
					log.Error("send raw transaction fail", "err", err)
					return err
				}
				returnWithdrawsList[index].Hash = common.HexToHash(txHash)
				returnWithdrawsList[index].GUID = withdraw.GUID
				balanceItem := database.Balances{
					Address:      hotWallet.Address,
					TokenAddress: withdraw.TokenAddress,
					LockBalance:  withdraw.Amount,
				}
				balanceList = append(balanceList, balanceItem)
				index++
			}

			retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
			if _, err := retry.Do[interface{}](w.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
				if err := w.db.Transaction(func(tx *database.DB) error {
					// 将转出去的热钱包余额锁定
					err = w.db.Balances.UpdateBalances(balanceList, false)
					if err != nil {
						log.Error("mark withdraw send fail", "err", err)
						return err
					}

					err = w.db.Withdraws.MarkWithdrawsToSend(returnWithdrawsList)
					if err != nil {
						log.Error("mark withdraw send fail", "err", err)
						return err
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
		}
		return nil
	})
	return nil
}
