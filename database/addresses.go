package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math"

	"github.com/ethereum/go-ethereum/common"
)

type Addresses struct {
	GUID         uuid.UUID      `gorm:"primaryKey" json:"guid"`
	UserUid      string         `json:"user_uid"`
	Address      common.Address `json:"address" gorm:"serializer:bytes"`
	ToKenAddress common.Address `json:"token_address" gorm:"serializer:bytes"`
	AddressType  uint8          `json:"address_type"` //0:用户地址；1:热钱包地址(归集地址)；2:冷钱包地址
	PrivateKey   string         `json:"private_key"`
	PublicKey    string         `json:"public_key"`
	Balance      string         `json:"balance"`
	Timestamp    uint64
}

type AddressesView interface {
	QueryAddressesByToAddres(*common.Address) (*Addresses, error)
	QueryHotWalletInfo() (*Addresses, error)
	QueryColdWalletInfo() (*Addresses, error)
	UnCollectionList(decimal uint64) ([]Addresses, error)
}

type AddressesDB interface {
	AddressesView

	StoreAddressess([]Addresses, uint64) error
}

type addressesDB struct {
	gorm *gorm.DB
}

func (db *addressesDB) QueryAddressesByToAddres(address *common.Address) (*Addresses, error) {
	var addressEntry Addresses
	db.gorm.Table("addresses").Where("address", address).Take(&addressEntry)
	return &addressEntry, nil
}

func NewAddressesDB(db *gorm.DB) AddressesDB {
	return &addressesDB{gorm: db}
}

func (db *addressesDB) StoreAddressess(addressList []Addresses, addressLength uint64) error {
	result := db.gorm.CreateInBatches(&addressList, int(addressLength))
	return result.Error
}

func (db *addressesDB) QueryHotWalletInfo() (*Addresses, error) {
	var addressEntry Addresses
	db.gorm.Table("addresses").Where("address_type", 1).Take(&addressEntry)
	return &addressEntry, nil
}

func (db *addressesDB) QueryColdWalletInfo() (*Addresses, error) {
	var addressEntry Addresses
	db.gorm.Table("addresses").Where("address_type", 2).Take(&addressEntry)
	return &addressEntry, nil
}

func (db *addressesDB) UnCollectionList(decimal uint64) ([]Addresses, error) {
	var addressList []Addresses
	err := db.gorm.Table("addresses").Where("balance >=?", math.Pow(10, float64(decimal))).Find(&addressList).Error
	if err != nil {
		return nil, err
	}
	return addressList, nil
}
