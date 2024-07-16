package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Balances struct {
	GUID         uuid.UUID      `gorm:"primaryKey" json:"guid"`
	Address      common.Address `json:"address" gorm:"serializer:bytes"`
	ToKenAddress common.Address `json:"token_address" gorm:"serializer:bytes"`
	AddressType  uint8          `json:"address_type"` //0:用户地址；1:热钱包地址(归集地址)；2:冷钱包地址
	Balance      *big.Int       `gorm:"serializer:u256;column:balance" db:"balance" json:"Balance" form:"balance"`
	Timestamp    uint64
}

type BalancesView interface {
	QueryHotWalletBalance(address, tokenAddress common.Address) (*Balances, error)
	UnCollectionList(decimal uint64) ([]Balances, error)
	QueryHotWalletBalances() ([]Balances, error)
}

type BalancesDB interface {
	BalancesView

	UpdateOrCreate([]TokenBalance) error
}

type balancesDB struct {
	gorm *gorm.DB
}

func NewBalancesDB(db *gorm.DB) BalancesDB {
	return &balancesDB{gorm: db}
}

func (db *balancesDB) QueryBalancesByToAddres(address *common.Address) (*Balances, error) {
	var balanceEntry Balances
	db.gorm.Table("balances").Where("address", address.String()).Take(&balanceEntry)
	return &balanceEntry, nil
}

func (db *balancesDB) QueryHotWalletBalances() ([]Balances, error) {
	var balanceList []Balances
	err := db.gorm.Table("balances").Where("address_type = ?", 1).Find(&balanceList).Error
	if err != nil {
		return nil, err
	}
	return balanceList, nil
}

func (db *balancesDB) UnCollectionList(decimal uint64) ([]Balances, error) {
	var balanceList []Balances
	err := db.gorm.Table("balances").Where("balance >=?", 10^decimal).Find(&balanceList).Error
	if err != nil {
		return nil, err
	}
	return balanceList, nil
}

func (db *balancesDB) QueryHotWalletBalance(address, tokenAddress common.Address) (*Balances, error) {
	var balanceEntry Balances
	db.gorm.Table("balances").Where("address = ? and token_address = ?", address.String(), tokenAddress.String()).Take(&balanceEntry)
	return &balanceEntry, nil
}

func (db *balancesDB) UpdateOrCreate(balanceList []TokenBalance) error {
	var hotWalletBalance *Balances
	db.gorm.Table("balances").Where("address_type = ?", 1).Take(hotWalletBalance)
	for _, value := range balanceList {
		var userBalanceEntry *Balances
		db.gorm.Table("balances").Where("address = ? and token_address = ? and address_type = ?", value.Address, value.ToKenAddress, 0).Take(userBalanceEntry)
		if userBalanceEntry == nil {
			result := db.gorm.Create(&value)
			return result.Error
		} else {
			// 0:充值；1:提现；2:归集；3:热转冷；4:冷转热
			if value.TxType == 0 {
				userBalanceEntry.Balance = new(big.Int).Add(userBalanceEntry.Balance, value.Balance)
			} else if value.TxType == 1 {
				userBalanceEntry.Balance = new(big.Int).Sub(userBalanceEntry.Balance, value.Balance)
			} else if value.TxType == 2 {
				userBalanceEntry.Balance = new(big.Int).Sub(userBalanceEntry.Balance, value.Balance)
				hotWalletBalance.Balance = new(big.Int).Add(hotWalletBalance.Balance, value.Balance)
			} else if value.TxType == 3 {
				hotWalletBalance.Balance = new(big.Int).Sub(hotWalletBalance.Balance, value.Balance)
			} else if value.TxType == 4 {
				hotWalletBalance.Balance = new(big.Int).Add(hotWalletBalance.Balance, value.Balance)
			}
			if value.TxType == 0 || value.TxType == 1 || value.TxType == 2 {
				err := db.gorm.Save(&userBalanceEntry).Error
				if err != nil {
					return err
				}
			}
			if value.TxType == 2 || value.TxType == 3 || value.TxType == 4 {
				err := db.gorm.Save(&hotWalletBalance).Error
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
