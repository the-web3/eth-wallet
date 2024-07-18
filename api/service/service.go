package service

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"

	"github.com/the-web3/eth-wallet/api/models"
	"github.com/the-web3/eth-wallet/database"
)

type Service interface {
	GetDepositList(*models.QueryDWParams) (*models.DepositsResponse, error)
	GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error)
	SubmitWithdrawFromBusiness(params *models.SubmitDWParams) (*models.SubmitWithdrawsResponse, error)

	SubmitDWParams(fromAddress string, toAddress string, tokenAddress string, amount string) (*models.SubmitDWParams, error)
	QueryDWListParams(address string, page string, pageSize string, order string) (*models.QueryDWParams, error)
	QueryPageListParams(page string, pageSize string, order string) (*models.QueryPageParams, error)
}

type HandlerSvc struct {
	v             *Validator
	depositsView  database.DepositsView
	withdrawsView database.WithdrawsView
}

func New(v *Validator, dsv database.DepositsView, wdv database.WithdrawsView) Service {
	return &HandlerSvc{
		v:             v,
		depositsView:  dsv,
		withdrawsView: wdv,
	}
}

func (h HandlerSvc) GetDepositList(params *models.QueryDWParams) (*models.DepositsResponse, error) {
	addressToLower := strings.ToLower(params.Address)
	depositList, total := h.depositsView.ApiDepositList(addressToLower, params.Page, params.PageSize, params.Order)
	return &models.DepositsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: depositList,
	}, nil
}

func (h HandlerSvc) GetWithdrawalList(params *models.QueryDWParams) (*models.WithdrawsResponse, error) {
	addressToLower := strings.ToLower(params.Address)
	withdrawList, total := h.withdrawsView.ApiWithdrawList(addressToLower, params.Page, params.PageSize, params.Order)
	return &models.WithdrawsResponse{
		Current: params.Page,
		Size:    params.PageSize,
		Total:   total,
		Records: withdrawList,
	}, nil
}

func (h HandlerSvc) SubmitWithdrawFromBusiness(params *models.SubmitDWParams) (*models.SubmitWithdrawsResponse, error) {
	err := h.withdrawsView.SubmitWithdrawFromBusiness(params.FromAddress, params.ToAddress, params.ToAddress, params.Amount)
	if err != nil {
		return &models.SubmitWithdrawsResponse{
			Code: 4000,
			Msg:  "submit transaction fail",
		}, nil
	}
	return &models.SubmitWithdrawsResponse{
		Code: 2000,
		Msg:  "submit transaction success",
	}, nil
}

func (h HandlerSvc) SubmitDWParams(fromAddress string, toAddress string, tokenAddress string, amount string) (*models.SubmitDWParams, error) {
	fromAddr, err := h.v.ParseValidateAddress(fromAddress)
	if err != nil {
		log.Error("invalid address param", "address", fromAddr.String(), "err", err)
		return nil, err
	}

	toAddr, err := h.v.ParseValidateAddress(toAddress)
	if err != nil {
		log.Error("invalid address param", "address", toAddr.String(), "err", err)
		return nil, err
	}

	tokenAddr, err := h.v.ParseValidateAddress(tokenAddress)
	if err != nil {
		log.Error("invalid address param", "address", toAddr.String(), "err", err)
		return nil, err
	}

	transferAmount, _ := new(big.Int).SetString(amount, 10)
	if transferAmount.Cmp(big.NewInt(0)) < 0 {
		log.Error("invalid amount param", "transferAmount", transferAmount.String(), "err", err)
		return nil, err
	}

	return &models.SubmitDWParams{
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		TokenAddress: tokenAddr,
		Amount:       transferAmount,
	}, nil
}

func (h HandlerSvc) QueryDWListParams(address string, page string, pageSize string, order string) (*models.QueryDWParams, error) {
	var paraAddress string
	if address == "0x00" {
		paraAddress = "0x00"
	} else {
		addr, err := h.v.ParseValidateAddress(address)
		if err != nil {
			log.Error("invalid address param", "address", address, "err", err)
			return nil, err
		}
		paraAddress = addr.String()
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}
	pageVal := h.v.ValidatePage(pageInt)

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, err
	}
	pageSizeVal := h.v.ValidatePageSize(pageSizeInt)
	orderBy := h.v.ValidateOrder(order)

	return &models.QueryDWParams{
		Address:  paraAddress,
		Page:     pageVal,
		PageSize: pageSizeVal,
		Order:    orderBy,
	}, nil
}

func (h HandlerSvc) QueryPageListParams(page string, pageSize string, order string) (*models.QueryPageParams, error) {
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}
	pageValue := h.v.ValidatePage(pageInt)
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, err
	}
	pageSizeValue := h.v.ValidatePageSize(pageSizeInt)
	orderBy := h.v.ValidateOrder(order)
	return &models.QueryPageParams{
		Page:     pageValue,
		PageSize: pageSizeValue,
		Order:    orderBy,
	}, nil
}
