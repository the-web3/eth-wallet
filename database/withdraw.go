package database

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Withdraw struct {
	GUID             uuid.UUID      `gorm:"primaryKey" json:"guid"`
	BlockHash        common.Hash    `gorm:"column:block_hash;serializer:bytes"  db:"block_hash" json:"block_hash"`
	BlockNumber      *big.Int       `gorm:"serializer:u256;column:block_number" db:"block_number" json:"BlockNumber" form:"block_number"`
	Hash             common.Hash    `gorm:"column:hash;serializer:bytes"  db:"hash" json:"hash"`
	FromAddress      common.Address `json:"from_address" gorm:"serializer:bytes"`
	ToAddress        common.Address `json:"to_address" gorm:"serializer:bytes"`
	ToKenAddress     common.Address `json:"token_address" gorm:"serializer:bytes"`
	Fee              *big.Int       `gorm:"serializer:u256;column:fee" db:"fee" json:"Fee" form:"fee"`
	Amount           *big.Int       `gorm:"serializer:u256;column:amount" db:"amount" json:"Amount" form:"amount"`
	Status           uint8          `json:"status"` // 0:提现未签名发送,1:提现已经发送到区块链网络；2:提现已上链；3:提现在钱包层已完成；4:提现已通知业务；5:提现成功
	TransactionIndex *big.Int       `gorm:"serializer:u256;column:transaction_index" db:"transaction_index" json:"TransactionIndex" form:"transaction_index"`
	TxSignHex        string         `json:"tx_sign_hex"`
	Timestamp        uint64
}

type WithdrawView interface {
	QueryWithdrawByHash(hash common.Hash) (*Withdraw, error)
	UnSendWithdrawList() ([]Withdraw, error)
}

type WithdrawDB interface {
	WithdrawView

	StoreWithdraws([]Withdraw, uint64) error
	UpdateTransactionStatus(withdrawList []Withdraw) error
	MarkWithdrawToSend(withdrawList []Withdraw) error
}

type withdrawDB struct {
	gorm *gorm.DB
}

func (db *withdrawDB) QueryWithdrawByHash(hash common.Hash) (*Withdraw, error) {
	var withdrawEntity Withdraw
	db.gorm.Table("withdraw").Where("hash", hash.String()).Take(&withdrawEntity)
	return &withdrawEntity, nil
}

func (db *withdrawDB) UpdateTransactionStatus(withdrawList []Withdraw) error {
	for i := 0; i < len(withdrawList); i++ {
		var withdrawSingle = Withdraw{}

		result := db.gorm.Where(&Transactions{Hash: withdrawList[i].Hash}).Take(&withdrawSingle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		withdrawSingle.Status = 2
		withdrawSingle.Fee = withdrawList[i].Fee
		err := db.gorm.Save(&withdrawSingle).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func NewWithdrawDB(db *gorm.DB) WithdrawDB {
	return &withdrawDB{gorm: db}
}

func (db *withdrawDB) StoreWithdraws(withdrawList []Withdraw, withdrawLength uint64) error {
	result := db.gorm.CreateInBatches(&withdrawList, int(withdrawLength))
	return result.Error
}

func (db *withdrawDB) UnSendWithdrawList() ([]Withdraw, error) {
	var withdrawList []Withdraw
	err := db.gorm.Table("withdraw").Where("status = ?", 0).Find(&withdrawList).Error
	if err != nil {
		return nil, err
	}
	return withdrawList, nil
}

func (db *withdrawDB) MarkWithdrawToSend(withdrawList []Withdraw) error {
	for i := 0; i < len(withdrawList); i++ {
		var withdrawSingle = Withdraw{}
		result := db.gorm.Where(&Transactions{GUID: withdrawList[i].GUID}).Take(&withdrawSingle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		withdrawSingle.Hash = withdrawList[i].Hash
		withdrawSingle.Status = 1
		err := db.gorm.Save(&withdrawSingle).Error
		if err != nil {
			return err
		}
	}
	return nil
}
