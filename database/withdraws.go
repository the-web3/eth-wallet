package database

import (
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Withdraws struct {
	GUID             uuid.UUID      `gorm:"primaryKey" json:"guid"`
	BlockHash        common.Hash    `gorm:"column:block_hash;serializer:bytes"  db:"block_hash" json:"block_hash"`
	BlockNumber      *big.Int       `gorm:"serializer:u256;column:block_number" db:"block_number" json:"BlockNumber" form:"block_number"`
	Hash             common.Hash    `gorm:"column:hash;serializer:bytes"  db:"hash" json:"hash"`
	FromAddress      common.Address `json:"from_address" gorm:"serializer:bytes;column:from_address"`
	ToAddress        common.Address `json:"to_address" gorm:"serializer:bytes;column:to_address"`
	TokenAddress     common.Address `json:"token_address" gorm:"serializer:bytes;column:token_address"`
	Fee              *big.Int       `gorm:"serializer:u256;column:fee" db:"fee" json:"Fee" form:"fee"`
	Amount           *big.Int       `gorm:"serializer:u256;column:amount" db:"amount" json:"Amount" form:"amount"`
	Status           uint8          `json:"status"` // 0:提现未签名发送,1:提现已经发送到区块链网络；2:提现已上链；3:提现在钱包层已完成；4:提现已通知业务；5:提现成功
	TransactionIndex *big.Int       `gorm:"serializer:u256;column:transaction_index" db:"transaction_index" json:"TransactionIndex" form:"transaction_index"`
	TxSignHex        string         `json:"tx_sign_hex" gorm:"column:tx_sign_hex"`
	Timestamp        uint64
}

type WithdrawsView interface {
	QueryWithdrawsByHash(hash common.Hash) (*Withdraws, error)
	UnSendWithdrawsList() ([]Withdraws, error)
	ApiWithdrawList(string, int, int, string) ([]Withdraws, int64)

	SubmitWithdrawFromBusiness(fromAddress common.Address, toAddress common.Address, TokenAddress common.Address, amount *big.Int) error
}

type WithdrawsDB interface {
	WithdrawsView

	StoreWithdraws([]Withdraws, uint64) error
	UpdateTransactionStatus(withdrawsList []Withdraws) error
	MarkWithdrawsToSend(withdrawsList []Withdraws) error
}

type withdrawsDB struct {
	gorm *gorm.DB
}

func (db *withdrawsDB) ApiWithdrawList(address string, page int, pageSize int, order string) (withdraws []Withdraws, total int64) {
	var totalRecord int64
	var withdrawList []Withdraws
	queryStateRoot := db.gorm.Table("withdraws")
	if address != "0x00" {
		err := db.gorm.Table("withdraws").Select("block_number").Where("from_address = ?", address).Count(&totalRecord).Error
		if err != nil {
			log.Error("get withdraws list by address count fail")
		}
		queryStateRoot.Where(" from_address = ?", address).Offset((page - 1) * pageSize).Limit(pageSize)
	} else {
		err := db.gorm.Table("withdraws").Select("block_number").Count(&totalRecord).Error
		if err != nil {
			log.Error("get withdraws list by address count fail ")
		}
		queryStateRoot.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if strings.ToLower(order) == "asc" {
		queryStateRoot.Order("timestamp asc")
	} else {
		queryStateRoot.Order("timestamp desc")
	}
	qErr := queryStateRoot.Find(&withdrawList).Error
	if qErr != nil {
		log.Error("get withdraws list fail", "err", qErr)
	}
	return withdrawList, totalRecord
}

func (db *withdrawsDB) QueryWithdrawsByHash(hash common.Hash) (*Withdraws, error) {
	var withdrawsEntity Withdraws
	result := db.gorm.Table("withdraws").Where("hash", hash.String()).Take(&withdrawsEntity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &withdrawsEntity, nil
}

func (db *withdrawsDB) SubmitWithdrawFromBusiness(fromAddress common.Address, toAddress common.Address, TokenAddress common.Address, amount *big.Int) error {
	withdrawS := Withdraws{
		GUID:             uuid.New(),
		BlockHash:        common.Hash{},
		BlockNumber:      big.NewInt(1),
		Hash:             common.Hash{},
		FromAddress:      fromAddress,
		ToAddress:        toAddress,
		TokenAddress:     TokenAddress,
		Fee:              big.NewInt(1),
		Amount:           amount,
		Status:           0,
		TransactionIndex: big.NewInt(time.Now().Unix()),
		TxSignHex:        "",
		Timestamp:        uint64(time.Now().Unix()),
	}
	errC := db.gorm.Create(withdrawS).Error
	if errC != nil {
		log.Error("create withdraw fail", "err", errC)
		return errC
	}
	return nil
}

func (db *withdrawsDB) UpdateTransactionStatus(withdrawsList []Withdraws) error {
	for i := 0; i < len(withdrawsList); i++ {
		var withdrawsSingle = Withdraws{}

		result := db.gorm.Where(&Transactions{Hash: withdrawsList[i].Hash}).Take(&withdrawsSingle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		withdrawsSingle.Status = 2
		withdrawsSingle.Fee = withdrawsList[i].Fee
		err := db.gorm.Save(&withdrawsSingle).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func NewWithdrawsDB(db *gorm.DB) WithdrawsDB {
	return &withdrawsDB{gorm: db}
}

func (db *withdrawsDB) StoreWithdraws(withdrawsList []Withdraws, withdrawsLength uint64) error {
	result := db.gorm.CreateInBatches(&withdrawsList, int(withdrawsLength))
	return result.Error
}

func (db *withdrawsDB) UnSendWithdrawsList() ([]Withdraws, error) {
	var withdrawsList []Withdraws
	err := db.gorm.Table("withdraws").Where("status = ?", 0).Find(&withdrawsList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return withdrawsList, nil
}

func (db *withdrawsDB) MarkWithdrawsToSend(withdrawsList []Withdraws) error {
	for i := 0; i < len(withdrawsList); i++ {
		var withdrawsSingle = Withdraws{}
		result := db.gorm.Where(&Transactions{GUID: withdrawsList[i].GUID}).Take(&withdrawsSingle)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		withdrawsSingle.Hash = withdrawsList[i].Hash
		withdrawsSingle.Status = 1
		err := db.gorm.Save(&withdrawsSingle).Error
		if err != nil {
			return err
		}
	}
	return nil
}
