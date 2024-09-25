package ethereum

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestCreateAddressByKeyPairs(t *testing.T) {
	address, err := CreateAddressByKeyPairs()
	if err != nil {
		return
	}
	fmt.Println(address.PrivateKey, address.PublicKey, address.Address)
}

func TestCreateAddressFromPrivateKey(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	priKeyHex, address, err := CreateAddressFromPrivateKey(privKey)
	if err != nil {
		t.Fatalf("Error creating address from private key: %v", err)
	}

	expectedPriKeyHex := hex.EncodeToString(privKey.D.Bytes())
	if priKeyHex != expectedPriKeyHex {
		t.Errorf("Expected private key hex: %s, got: %s", expectedPriKeyHex, priKeyHex)
	}

	expectedAddress := crypto.PubkeyToAddress(privKey.PublicKey).String()
	if address != expectedAddress {
		t.Errorf("Expected address: %s, got: %s", expectedAddress, address)
	}
}

func TestPublicKeyToAddress(t *testing.T) {
	// Generate a new ECDSA private key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Encode the public key in uncompressed format
	pubKeyBytes := elliptic.Marshal(crypto.S256(), privKey.PublicKey.X, privKey.PublicKey.Y)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	// Calculate the expected address using common.BytesToAddress
	hash := crypto.Keccak256(pubKeyBytes[1:])
	expectedAddress := common.BytesToAddress(hash[12:]).String()

	t.Logf("Public Key: %s", pubKeyHex)
	t.Logf("Expected Address: %s", expectedAddress)

	// Call the function PublicKeyToAddress
	gotAddress, err := PublicKeyToAddress(pubKeyHex)
	if err != nil {
		t.Fatalf("Error creating address from public key: %v", err)
	}

	t.Logf("Got Address: %s", gotAddress)

	// Compare the expected address with the address generated by the function
	if gotAddress != expectedAddress {
		t.Errorf("Expected address: %s, got: %s", expectedAddress, gotAddress)
	}
}
