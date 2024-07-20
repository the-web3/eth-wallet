package database

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TokenBalance struct {
	Address      common.Address `json:"address"`
	TokenAddress common.Address `json:"to_ken_address"`
	Balance      *big.Int       `json:"balance"`
	LockBalance  *big.Int       `json:"lock_balance"`
	TxType       uint8          `json:"tx_type"` // 0:充值；1:提现；2:归集；3:热转冷；4:冷转热
}
