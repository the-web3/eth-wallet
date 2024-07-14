package database

import (
	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
	common2 "github.com/the-web3/eth-wallet/database/utils"
)

type Token struct {
	GUID         uuid.UUID      `gorm:"primaryKey" json:"guid"`
	ToKenAddress common.Address `json:"token_address" gorm:"serializer:bytes"`
	Uint         uint8          `json:"uint"`
	TokenName    string         `json:"token_name"`
	Timestamp    uint64
}

type TokenView interface {
	TokenInfoByAddress(common.Address) (*Token, error)
}

type TokenDB interface {
	TokenView

	StoreTokens([]Token, uint64) error
}

type tokenDB struct {
	gorm *gorm.DB
}

func NewTokenDB(db *gorm.DB) TokenDB {
	return &tokenDB{gorm: db}
}

func (db *tokenDB) StoreTokens(headers []Token, blockLength uint64) error {
	result := db.gorm.CreateInBatches(&headers, common2.BatchInsertSize)
	return result.Error
}

func (db *tokenDB) TokenInfoByAddress(address common.Address) (*Token, error) {
	var tokenEntry Token
	db.gorm.Table("addresses").Where("token_address", address).Take(&tokenEntry)
	return &tokenEntry, nil
}
