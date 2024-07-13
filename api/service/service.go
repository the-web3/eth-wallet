package service

import (
	"github.com/the-web3/eth-wallet/api/models"
)

type Service interface {
	GetDepositList(*models.QueryDWParams) (*models.DepositsResponse, error)
	GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error)
}

type HandlerSvc struct {
	v *Validator
}

func New(v *Validator) Service {
	return &HandlerSvc{
		v: v,
	}
}

func (h HandlerSvc) GetDepositList(params *models.QueryDWParams) (*models.DepositsResponse, error) {
	return &models.DepositsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   100,
	}, nil
}

func (h HandlerSvc) GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error) {
	return &models.WithdrawsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   100,
	}, nil
}
