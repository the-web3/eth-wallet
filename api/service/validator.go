package service

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type Validator struct{}

func (v *Validator) ParseValidateAddress(addr string) (common.Address, error) {
	var parsedAddr common.Address
	if addr != "0x00" {
		if !common.IsHexAddress(addr) {
			return common.Address{}, errors.New("address must be represented as a valid hexadecimal string")
		}
		parsedAddr = common.HexToAddress(addr)
		if parsedAddr == common.HexToAddress("0x0") {
			return common.Address{}, errors.New("address cannot be the zero address")
		}
	}
	return parsedAddr, nil
}

func (v *Validator) ValidatePage(page int) int {
	var validPage int
	if page <= 0 {
		validPage = 1
	} else {
		validPage = page
	}
	return validPage
}

func (v *Validator) ValidatePageSize(pageSize int) int {
	var validPageSize int
	if pageSize <= 0 || pageSize > 1000 {
		validPageSize = 20
	} else {
		validPageSize = pageSize
	}
	return validPageSize
}

func (v *Validator) ValidateOrder(order string) string {
	if order == "asc" || order == "ASC" || order == "DESC" || order == "desc" {
		return order
	} else {
		return "desc"
	}
}

func (v *Validator) ValidateIdOrIndex(idOrIndex uint64) error {
	if idOrIndex <= 0 {
		return errors.New("page size must be more than 0")
	}
	return nil
}
