package tools

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"

	"github.com/the-web3/eth-wallet/database"
	"github.com/the-web3/eth-wallet/wallet/ethereum"
)

func CreateAddressTools(ctx *cli.Context, db *database.DB) error {
	var addressList []database.Addresses
	for index := 0; index < 100; index++ {
		addressStruct, err := ethereum.CreateAddressByKeyPairs()
		if err != nil {
			log.Error("create address error", err)
			return err
		}
		var AddressType uint8
		var UserUid string
		if index == 1 {
			AddressType = 1
			UserUid = "hot-wallet-for-the-web3"
		} else if index == 2 {
			AddressType = 2
			UserUid = "cold-wallet-for-the-web3"
		} else {
			UserUid = "useruid"
			AddressType = 0
		}
		addressItem := database.Addresses{
			GUID:        uuid.New(),
			UserUid:     UserUid,
			Address:     common.Address(common.FromHex(addressStruct.Address)),
			AddressType: AddressType,
			PrivateKey:  addressStruct.PrivateKey,
			PublicKey:   addressStruct.PublicKey,
			Timestamp:   uint64(index + 10000),
		}
		addressList = append(addressList, addressItem)
	}
	err := db.Addresses.StoreAddressess(addressList, uint64(len(addressList)))
	if err != nil {
		log.Error("store address error", err)
		return err
	}
	return nil
}
