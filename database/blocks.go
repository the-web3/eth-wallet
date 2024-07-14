package database

import (
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	common2 "github.com/the-web3/eth-wallet/database/utils"
)

type Blocks struct {
	Hash       common.Hash `gorm:"primaryKey;serializer:bytes"`
	ParentHash common.Hash `gorm:"serializer:bytes"`
	Number     *big.Int    `gorm:"serializer:u256"`
	Timestamp  uint64
	RLPHeader  *common2.RLPHeader `gorm:"serializer:rlp;column:rlp_bytes"`
}

func BlockHeaderFromHeader(header *types.Header) Blocks {
	return Blocks{
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Number:     header.Number,
		Timestamp:  header.Time,
		RLPHeader:  (*common2.RLPHeader)(header),
	}
}

type BlocksView interface {
	LatestBlocks() (*Blocks, error)
}

type BlocksDB interface {
	BlocksView

	StoreBlockss([]Blocks, uint64) error
}

type blocksDB struct {
	gorm *gorm.DB
}

func NewBlocksDB(db *gorm.DB) BlocksDB {
	return &blocksDB{gorm: db}
}

func (db *blocksDB) StoreBlockss(headers []Blocks, blockLength uint64) error {
	result := db.gorm.CreateInBatches(&headers, common2.BatchInsertSize)
	return result.Error
}

func (db *blocksDB) LatestBlocks() (*Blocks, error) {
	var l1Header Blocks
	result := db.gorm.Order("number DESC").Take(&l1Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Error("record not found for query latest block")
			return nil, nil
		}

		return nil, result.Error
	}

	return &l1Header, nil
}
