package database

import (
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Deposits struct {
	GUID             uuid.UUID      `gorm:"primaryKey" json:"guid"`
	BlockHash        common.Hash    `gorm:"column:block_hash;serializer:bytes"  db:"block_hash" json:"block_hash"`
	BlockNumber      *big.Int       `gorm:"serializer:u256;column:block_number" db:"block_number" json:"BlockNumber" form:"block_number"`
	Hash             common.Hash    `gorm:"column:hash;serializer:bytes"  db:"hash" json:"hash"`
	FromAddress      common.Address `json:"from_address" gorm:"serializer:bytes;column:from_address"`
	ToAddress        common.Address `json:"to_address" gorm:"serializer:bytes;column:to_address"`
	TokenAddress     common.Address `json:"token_address" gorm:"serializer:bytes;column:token_address"`
	Fee              *big.Int       `gorm:"serializer:u256;column:fee" db:"fee" json:"Fee" form:"fee"`
	Amount           *big.Int       `gorm:"serializer:u256;column:amount" db:"amount" json:"Amount" form:"amount"`
	Status           uint8          `json:"status"` //0:充值确认中,1:充值钱包层已到账；2:充值已通知业务层；3:充值完成
	TransactionIndex *big.Int       `gorm:"serializer:u256;column:transaction_index" db:"transaction_index" json:"TransactionIndex" form:"transaction_index"`
	Timestamp        uint64
}

type DepositsView interface {
	ApiDepositList(string, int, int, string) ([]Deposits, int64)
}

type DepositsDB interface {
	DepositsView

	StoreDeposits([]Deposits, uint64) error
	UpdateDepositsStatus(blockNumber uint64) error
}

type depositsDB struct {
	gorm *gorm.DB
}

func (db *depositsDB) ApiDepositList(address string, page int, pageSize int, order string) (l1l2List []Deposits, total int64) {
	var totalRecord int64
	var depositList []Deposits
	queryStateRoot := db.gorm.Table("deposits")
	if address != "0x00" {
		err := db.gorm.Table("deposits").Select("block_number").Where("to_address = ?", address).Count(&totalRecord).Error
		if err != nil {
			log.Error("get deposit list by address count fail")
		}
		queryStateRoot.Where(" to_address = ?", address).Offset((page - 1) * pageSize).Limit(pageSize)
	} else {
		err := db.gorm.Table("deposits").Select("block_number").Count(&totalRecord).Error
		if err != nil {
			log.Error("get deposit list by address count fail ")
		}
		queryStateRoot.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if strings.ToLower(order) == "asc" {
		queryStateRoot.Order("timestamp asc")
	} else {
		queryStateRoot.Order("timestamp desc")
	}
	qErr := queryStateRoot.Find(&depositList).Error
	if qErr != nil {
		log.Error("get deposit list fail", "err", qErr)
	}
	return depositList, totalRecord
}

func (db *depositsDB) UpdateDepositsStatus(blockNumber uint64) error {
	result := db.gorm.Model(&Deposits{}).Where("status = ? and block_number <= ?", 0, blockNumber).Updates(map[string]interface{}{"status": 1})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	return nil
}

func NewDepositsDB(db *gorm.DB) DepositsDB {
	return &depositsDB{gorm: db}
}

func (db *depositsDB) StoreDeposits(depositList []Deposits, depositLength uint64) error {
	result := db.gorm.CreateInBatches(&depositList, int(depositLength))
	if result.Error != nil {
		log.Error("create deposit batch fail", "Err", result.Error)
		return result.Error
	}
	return nil
}
