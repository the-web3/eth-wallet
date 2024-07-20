package eth_wallet

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"github.com/the-web3/eth-wallet/config"
	"github.com/the-web3/eth-wallet/database"
	"github.com/the-web3/eth-wallet/wallet"
	"github.com/the-web3/eth-wallet/wallet/node"
	"sync/atomic"
)

type EthWallet struct {
	deposit        *wallet.Deposit
	withdraw       *wallet.Withdraw
	collectionCold *wallet.CollectionCold

	shutdown context.CancelCauseFunc
	stopped  atomic.Bool
}

func NewEthWallet(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*EthWallet, error) {
	ethClient, err := node.DialEthClient(ctx, cfg.Chain.RpcUrl)
	if err != nil {
		return nil, err
	}

	db, err := database.NewDB(ctx, cfg.MasterDB)
	if err != nil {
		log.Error("init database fail", err)
		return nil, err
	}

	deposit, _ := wallet.NewDeposit(cfg, db, ethClient, shutdown)
	withdraw, _ := wallet.NewWithdraw(cfg, db, ethClient, shutdown)
	collectionCold, _ := wallet.NewCollectionCold(cfg, db, ethClient, shutdown)

	out := &EthWallet{
		deposit:        deposit,
		withdraw:       withdraw,
		collectionCold: collectionCold,
		shutdown:       shutdown,
	}

	return out, nil
}

func (ew *EthWallet) Start(ctx context.Context) error {
	err := ew.deposit.Start()
	if err != nil {
		return err
	}
	//err = ew.withdraw.Start()
	//if err != nil {
	//	return err
	//}
	err = ew.collectionCold.Start()
	if err != nil {
		return err
	}
	return nil
}

func (ew *EthWallet) Stop(ctx context.Context) error {
	err := ew.deposit.Close()
	if err != nil {
		return err
	}
	//err = ew.withdraw.Close()
	//if err != nil {
	//	return err
	//}

	err = ew.collectionCold.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ew *EthWallet) Stopped() bool {
	return ew.stopped.Load()
}
