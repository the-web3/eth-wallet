package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	"github.com/the-web3/eth-wallet/wallet/ethereum"
	"github.com/the-web3/eth-wallet/wallet/node"
	"github.com/the-web3/eth-wallet/wallet/retry"
)

type CollectionCold struct {
	db             *database.DB
	chainConf      *config.ChainConfig
	client         node.EthClient
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
}

func NewCollectionCold(cfg *config.Config, db *database.DB, client node.EthClient, shutdown context.CancelCauseFunc) (*CollectionCold, error) {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &CollectionCold{
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

func (cc *CollectionCold) Close() error {
	var result error
	cc.resourceCancel()
	if err := cc.tasks.Wait(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to await deposit %w"), err)
	}
	return nil
}

func (cc *CollectionCold) Start() error {
	log.Info("start collection and cold......")
	tickerCollectionColdWorker := time.NewTicker(time.Second * 5)
	cc.tasks.Go(func() error {
		for range tickerCollectionColdWorker.C {
			err := cc.Collection()
			if err != nil {
				log.Error("collect fail", "err", err)
				return err
			}
		}
		return nil
	})

	cc.tasks.Go(func() error {
		for range tickerCollectionColdWorker.C {
			err := cc.ToCold()
			if err != nil {
				log.Error("to cold fail", "err", err)
				return err
			}
		}
		return nil
	})

	return nil
}

func (cc *CollectionCold) ToCold() error {
	hotWalletInfo, err := cc.db.Addresses.QueryHotWalletInfo()
	if err != nil {
		log.Error("query hot wallet info err", "err", err)
		return err
	}
	var txList []database.Transactions
	balance, _ := strconv.ParseFloat(hotWalletInfo.Balance, 64)

	if balance >= math.Pow(10, 18) {
		coldWalletInfo, err := cc.db.Addresses.QueryColdWalletInfo()
		if err != nil {
			log.Error("query cold wallet info err", "err", err)
			return err
		}

		// nonce
		nonce, err := cc.client.TxCountByAddress(hotWalletInfo.Address)
		if err != nil {
			log.Error("query nonce by address fail", "err", err)
			return err
		}

		var buildData []byte
		var gasLimit uint64
		var toAddress *common.Address
		var amount *big.Int
		if hotWalletInfo.ToKenAddress.Hex() != "0x00" {
			buildData = ethereum.BuildErc20Data(coldWalletInfo.Address, big.NewInt(int64(balance)))
			toAddress = &hotWalletInfo.ToKenAddress
			gasLimit = TokenGasLimit
			amount = big.NewInt(0)
		} else {
			toAddress = &hotWalletInfo.Address
			gasLimit = EthGasLimit
			amount = big.NewInt(int64(balance))
		}
		dFeeTx := &types.DynamicFeeTx{
			ChainID:   big.NewInt(int64(cc.chainConf.ChainID)),
			Nonce:     nonce.Uint64(),
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       gasLimit,
			To:        toAddress,
			Value:     amount,
			Data:      buildData,
		}
		rawTx, err := ethereum.OfflineSignTx(dFeeTx, hotWalletInfo.PrivateKey, big.NewInt(int64(cc.chainConf.ChainID)))
		if err != nil {
			log.Error("offline transaction fail", "err", err)
			return err
		}
		//  sendRawTx
		log.Info("Offline sign tx success", "rawTx", rawTx)
		hash, err := cc.client.SendRawTransaction(rawTx)
		if err != nil {
			log.Error("send raw transaction fail", "err", err)
			return err
		}

		guid, _ := uuid.NewUUID()
		coldTx := database.Transactions{
			GUID:             guid,
			BlockHash:        common.Hash{},
			BlockNumber:      nil,
			Hash:             *hash,
			FromAddress:      hotWalletInfo.Address,
			ToAddress:        coldWalletInfo.Address,
			ToKenAddress:     hotWalletInfo.ToKenAddress,
			Fee:              "0",
			Amount:           strconv.FormatInt(int64(balance), 10),
			Status:           0,
			TxType:           2,
			TransactionIndex: nil,
			R:                "r",
			S:                "s",
			V:                "v",
			Timestamp:        1000,
		}
		txList = append(txList, coldTx)
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](cc.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := cc.db.Transaction(func(tx *database.DB) error {
			if err := tx.Transactions.StoreTransactions(txList, uint64(len(txList))); err != nil {
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
	return nil
}

func (cc *CollectionCold) Collection() error {
	uncollectionList, err := cc.db.Addresses.UnCollectionList(18)
	if err != nil {
		log.Error("query hot wallet info err", "err", err)
		return err
	}

	hotWalletInfo, err := cc.db.Addresses.QueryHotWalletInfo()
	if err != nil {
		log.Error("query hot wallet info err", "err", err)
		return err
	}
	var txList []database.Transactions
	for _, uncollect := range uncollectionList {
		// nonce
		nonce, err := cc.client.TxCountByAddress(uncollect.Address)
		if err != nil {
			log.Error("query nonce by address fail", "err", err)
			return err
		}

		var buildData []byte
		var gasLimit uint64
		var toAddress *common.Address
		var amount *big.Int
		amountToken, _ := strconv.ParseInt(uncollect.Balance, 10, 64)
		if uncollect.ToKenAddress.Hex() != "0x00" {
			buildData = ethereum.BuildErc20Data(hotWalletInfo.Address, big.NewInt(amountToken))
			toAddress = &uncollect.ToKenAddress
			gasLimit = TokenGasLimit
			amount = big.NewInt(0)
		} else {
			toAddress = &hotWalletInfo.Address
			gasLimit = EthGasLimit
			amount = big.NewInt(amountToken)
		}
		dFeeTx := &types.DynamicFeeTx{
			ChainID:   big.NewInt(int64(cc.chainConf.ChainID)),
			Nonce:     nonce.Uint64(),
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       gasLimit,
			To:        toAddress,
			Value:     amount,
			Data:      buildData,
		}
		rawTx, err := ethereum.OfflineSignTx(dFeeTx, uncollect.PrivateKey, big.NewInt(int64(cc.chainConf.ChainID)))
		if err != nil {
			log.Error("offline transaction fail", "err", err)
			return err
		}
		//  sendRawTx
		log.Info("Offline sign tx success", "rawTx", rawTx)

		hash, err := cc.client.SendRawTransaction(rawTx)
		if err != nil {
			log.Error("send raw transaction fail", "err", err)
			return err
		}
		guid, _ := uuid.NewUUID()
		collection := database.Transactions{
			GUID:             guid,
			BlockHash:        common.Hash{},
			BlockNumber:      nil,
			Hash:             *hash,
			FromAddress:      uncollect.Address,
			ToAddress:        hotWalletInfo.Address,
			ToKenAddress:     uncollect.ToKenAddress,
			Fee:              "0",
			Amount:           strconv.FormatInt(amountToken, 10),
			Status:           0,
			TxType:           2,
			TransactionIndex: nil,
			R:                "r",
			S:                "s",
			V:                "v",
			Timestamp:        1000,
		}
		txList = append(txList, collection)
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](cc.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := cc.db.Transaction(func(tx *database.DB) error {
			if err := tx.Transactions.StoreTransactions(txList, uint64(len(txList))); err != nil {
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
	return nil
}
