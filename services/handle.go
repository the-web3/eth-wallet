package services

import (
	"context"
	"strconv"

	"github.com/ethereum/go-ethereum/log"

	"github.com/the-web3/eth-wallet/proto/wallet"
)

func (s *RpcServer) SubmitWithdrawInfo(ctx context.Context, in *wallet.WithdrawReq) (*wallet.WithdrawRep, error) {
	log.Info("submit withdraw from address", "FromAddress", in.FromAddress)
	log.Info("submit withdraw to address", "ToAddress", in.ToAddress)
	log.Info("submit withdraw amount", "Amount", in.Amount)
	return &wallet.WithdrawRep{
		Code: strconv.Itoa(200),
		Msg:  "success request",
		Hash: "0x00000000000000",
	}, nil
}
