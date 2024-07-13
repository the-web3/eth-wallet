package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
)

type Addresses struct {
	GUID        uuid.UUID      `gorm:"primaryKey" json:"guid"`
	UserUid     string         `json:"user_uid"`
	Address     common.Address `json:"address" gorm:"serializer:bytes"`
	AddressType uint8          `json:"address_type"` //0:用户地址；1:热钱包地址(归集地址)；2:冷钱包地址
	PrivateKey  string         `json:"private_key"`
	PublicKey   string         `json:"public_key"`
	Balance     string         `json:"balance"`
	Timestamp   uint64
}

type AddressesView interface {
	Addresses(common.Hash) (*Addresses, error)
	AddressesWithFilter(Addresses) (*Addresses, error)
	AddressesWithScope(func(db *gorm.DB) *gorm.DB) (*Addresses, error)
	LatestAddresses() (*Addresses, error)
}

type AddressesDB interface {
	AddressesView

	StoreAddressess([]Addresses, uint64) error
}

type addressesDB struct {
	gorm *gorm.DB
}

func NewAddressesDB(db *gorm.DB) AddressesDB {
	return &addressesDB{gorm: db}
}

func (db *addressesDB) StoreAddressess(addressList []Addresses, addressLength uint64) error {
	result := db.gorm.CreateInBatches(&addressList, int(addressLength))
	return result.Error
}

func (db *addressesDB) Addresses(hash common.Hash) (*Addresses, error) {
	//TODO implement me
	panic("implement me")
}

func (db *addressesDB) AddressesWithFilter(addresses Addresses) (*Addresses, error) {
	//TODO implement me
	panic("implement me")
}

func (db *addressesDB) AddressesWithScope(f func(db *gorm.DB) *gorm.DB) (*Addresses, error) {
	//TODO implement me
	panic("implement me")
}

func (db *addressesDB) LatestAddresses() (*Addresses, error) {
	//TODO implement me
	panic("implement me")
}
