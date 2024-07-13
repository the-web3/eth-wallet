package eth_wallet

import (
	"context"
	"github.com/the-web3/eth-wallet/config"
	"github.com/the-web3/eth-wallet/wallet"
	"github.com/the-web3/eth-wallet/wallet/node"
	"sync/atomic"
)

type EthWallet struct {
	ethClient node.EthClient

	deposit        *wallet.Deposit
	withdraw       *wallet.Withdraw
	collectionCold *wallet.CollectionCold

	shutdown context.CancelCauseFunc
	stopped  atomic.Bool
}

func NewEthWallet(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*EthWallet, error) {
	ethClient, err := node.DialEthClient(ctx, "https://eth-mainnet.g.alchemy.com/v2/XZw9s8EsSyUtwDOjtVvzwL8N0T96Zxt0")
	if err != nil {
		return nil, err
	}

	deposit, _ := wallet.NewDeposit(ethClient, shutdown)
	withdraw, _ := wallet.NewWithdraw(ethClient, shutdown)
	collectionCold, _ := wallet.NewCollectionCold(ethClient, shutdown)

	out := &EthWallet{
		ethClient:      ethClient,
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
	err = ew.withdraw.Start()
	if err != nil {
		return err
	}
	err = ew.collectionCold.Start()
	if err != nil {
		return err
	}
	return nil
}

func (ew *EthWallet) Stop(ctx context.Context) error {
	return nil
}

func (ew *EthWallet) Stopped() bool {
	return ew.stopped.Load()
}
