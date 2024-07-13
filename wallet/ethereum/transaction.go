package ethereum

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"
)

func BuildErc20Data(toAddress common.Address, amount *big.Int) []byte {
	var data []byte

	tranferFnSignature := []byte("transfer(address,uint256)")
	hash := crypto.Keccak256Hash(tranferFnSignature)
	methodId := hash[:5]
	dataAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataAmount := common.LeftPadBytes(amount.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataAddress...)
	data = append(data, dataAmount...)

	return data
}

func BuildErc721Data(fromAddress, toAddress common.Address, tokenId *big.Int) []byte {
	var data []byte

	tranferFnSignature := []byte("safeTransferFrom(address,address,uint256)")
	hash := crypto.Keccak256Hash(tranferFnSignature)
	methodId := hash[:5]

	dataFromAddress := common.LeftPadBytes(fromAddress.Bytes(), 32)
	dataToAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataTokenId := common.LeftPadBytes(tokenId.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataFromAddress...)
	data = append(data, dataToAddress...)
	data = append(data, dataTokenId...)

	return data
}

func OfflineSignTx(txData *types.DynamicFeeTx, privateKey string, chainId *big.Int) (string, error) {
	privateKeyEcdsa, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)

	messegeHash := signer.Hash(tx)
	fmt.Println("messegeHash::", messegeHash)

	seckey := math.PaddedBigBytes(privateKeyEcdsa.D, privateKeyEcdsa.Params().BitSize/8)
	fmt.Println("seckey::", seckey)

	// =====cloadHsm====
	signature, _ := secp256k1.Sign(messegeHash[:], seckey)
	// =====cloadHsm====

	fmt.Println("signature::", signature)

	signedTxCloadHsm, _ := tx.WithSignature(signer, signature)

	signedCloadHsmTxData, err := rlp.EncodeToBytes(signedTxCloadHsm)

	fmt.Println("ssssaawwweeww=====", "0x"+hex.EncodeToString(signedCloadHsmTxData)[6:])

	signedTx, err := types.SignTx(tx, signer, privateKeyEcdsa)

	if err != nil {
		return "", err
	}

	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(signedTxData)[6:], nil
}
