package services

import (
	"context"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/the-web3/eth-wallet/proto/wallet"
)

func (s *RpcServer) SubmitWithdrawInfo(ctx context.Context, in *wallet.WithdrawReq) (*wallet.WithdrawRep, error) {
	log.Info("submit withdraw start....")
	amountBig := new(big.Int)
	_, ok := amountBig.SetString(in.Amount, 10)
	if !ok {
		log.Error("invalid input amount")
		return &wallet.WithdrawRep{
			Code: strconv.Itoa(4000),
			Msg:  "submit withdraw fail",
			Hash: common.Hash{}.String(),
		}, nil
	}
	amountBig.SetString(in.Amount, 10)
	err := s.db.Withdraws.SubmitWithdrawFromBusiness(common.HexToAddress(in.FromAddress), common.HexToAddress(in.ToAddress), common.HexToAddress(in.TokenAddress), amountBig)
	if err != nil {
		log.Error("submit withdraw fail", "err", err)
		return &wallet.WithdrawRep{
			Code: strconv.Itoa(4000),
			Msg:  "submit withdraw fail",
			Hash: common.Hash{}.String(),
		}, nil
	}
	return &wallet.WithdrawRep{
		Code: strconv.Itoa(2000),
		Msg:  "submit withdraw success",
		Hash: common.Hash{}.String(),
	}, nil
}

func (s *RpcServer) VerifyAddress(ctx context.Context, in *wallet.RiskVerifyAddressReq) (*wallet.RiskVerifyAddressRep, error) {
	return &wallet.RiskVerifyAddressRep{
		Code:   strconv.Itoa(200),
		Msg:    "success request",
		Verify: true,
	}, nil
}

func (s *RpcServer) VerifyWithdrawSign(ctx context.Context, in *wallet.RiskWithdrawVerifyReq) (*wallet.RiskWithdrawVerifyRep, error) {
	return &wallet.RiskWithdrawVerifyRep{
		Code:   strconv.Itoa(200),
		Msg:    "success request",
		Verify: true,
	}, nil
}

func (s *RpcServer) VerifyRiskDOrWNotify(ctx context.Context, in *wallet.RiskDOrWNotifyVerifyReq) (*wallet.RiskDOrWNotifyVerifyRep, error) {
	return &wallet.RiskDOrWNotifyVerifyRep{
		Code:   strconv.Itoa(200),
		Msg:    "success request",
		Verify: true,
	}, nil
}
