package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Deposit struct {
	GUID             uuid.UUID      `gorm:"primaryKey" json:"guid"`
	BlockHash        common.Hash    `gorm:"column:block_hash;serializer:bytes"  db:"block_hash" json:"block_hash"`
	BlockNumber      *big.Int       `gorm:"serializer:u256;column:block_number" db:"block_number" json:"BlockNumber" form:"block_number"`
	Hash             common.Hash    `gorm:"column:hash;serializer:bytes"  db:"hash" json:"hash"`
	FromAddress      common.Address `json:"from_address" gorm:"serializer:bytes"`
	ToAddress        common.Address `json:"to_address" gorm:"serializer:bytes"`
	ToKenAddress     common.Address `json:"token_address" gorm:"serializer:bytes"`
	Fee              *big.Int       `gorm:"serializer:u256;column:fee" db:"fee" json:"Fee" form:"fee"`
	Amount           *big.Int       `gorm:"serializer:u256;column:amount" db:"amount" json:"Amount" form:"amount"`
	Status           uint8          `json:"status"` //0:充值确认中,1:充值钱包层已到账；2:充值已通知业务层；3:充值完成
	TransactionIndex *big.Int       `gorm:"serializer:u256;column:transaction_index" db:"transaction_index" json:"TransactionIndex" form:"transaction_index"`
	Timestamp        uint64
}

type DepositView interface {
}

type DepositDB interface {
	DepositView

	StoreDeposits([]Deposit, uint64) error
	UpdateDepositStatus(blockNumber uint64) error
}

type depositDB struct {
	gorm *gorm.DB
}

func (db *depositDB) UpdateDepositStatus(blockNumber uint64) error {
	result := db.gorm.Model(&Deposit{}).Where("status = ? and block_number <= ?", 0, blockNumber).Updates(map[string]interface{}{"status": gorm.Expr("GREATEST(1)")})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	return nil
}

func NewDepositDB(db *gorm.DB) DepositDB {
	return &depositDB{gorm: db}
}

func (db *depositDB) StoreDeposits(depositList []Deposit, depositLength uint64) error {
	result := db.gorm.CreateInBatches(&depositList, int(depositLength))
	return result.Error
}
