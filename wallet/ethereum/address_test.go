package ethereum

import (
	"fmt"
	"testing"
)

func TestCreateAddressByKeyPairs(t *testing.T) {
	address, err := CreateAddressByKeyPairs()
	if err != nil {
		return
	}
	fmt.Println(address.PrivateKey, address.PublicKey, address.Address)
}
