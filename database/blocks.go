package database

import (
	"errors"
	"math/big"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	common2 "github.com/the-web3/eth-wallet/database/utils"
)

type Blocks struct {
	Hash       common.Hash `gorm:"primaryKey;serializer:bytes"`
	ParentHash common.Hash `gorm:"serializer:bytes"`
	Number     *big.Int    `gorm:"serializer:u256"`
	Timestamp  uint64
	RLPHeader  *common2.RLPHeader `gorm:"serializer:rlp;column:rlp_bytes"`
}

func BlocksFromHeader(header *types.Header) Blocks {
	return Blocks{
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Number:     header.Number,
		Timestamp:  header.Time,
		RLPHeader:  (*common2.RLPHeader)(header),
	}
}

type BlocksView interface {
	Blocks(common.Hash) (*Blocks, error)
	BlocksWithFilter(Blocks) (*Blocks, error)
	BlocksWithScope(func(db *gorm.DB) *gorm.DB) (*Blocks, error)
	LatestBlocks() (*Blocks, error)
}

type BlocksDB interface {
	BlocksView

	StoreBlockss([]Blocks) error
}

type blocksDB struct {
	gorm *gorm.DB
}

func NewBlocksDB(db *gorm.DB) BlocksDB {
	return &blocksDB{gorm: db}
}

func (db *blocksDB) StoreBlockss(headers []Blocks) error {
	result := db.gorm.CreateInBatches(&headers, common2.BatchInsertSize)
	return result.Error
}

func (db *blocksDB) Blocks(hash common.Hash) (*Blocks, error) {
	return db.BlocksWithFilter(Blocks{Hash: hash})
}

func (db *blocksDB) BlocksWithFilter(filter Blocks) (*Blocks, error) {
	return db.BlocksWithScope(func(gorm *gorm.DB) *gorm.DB { return gorm.Where(&filter) })
}

func (db *blocksDB) BlocksWithScope(scope func(*gorm.DB) *gorm.DB) (*Blocks, error) {
	var l1Header Blocks
	result := db.gorm.Scopes(scope).Take(&l1Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l1Header, nil
}

func (db *blocksDB) LatestBlocks() (*Blocks, error) {
	var l1Header Blocks
	result := db.gorm.Order("number DESC").Take(&l1Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &l1Header, nil
}
