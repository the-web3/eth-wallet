package ethereum

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestOfflineSignTx(t *testing.T) {
	privateKeyHex := "0cbb2ff952da876c4779200c83f6b90d73ea85a8da82e06c2276a11499922720"
	nonce := uint64(56)
	toAddress := common.HexToAddress("0x35096AD62E57e86032a3Bb35aDaCF2240d55421D")
	amount := big.NewInt(10000000000000000)
	gasLimit := uint64(21000)
	maxPriorityFeePerGas := big.NewInt(2600000000)
	maxFeePerGas := big.NewInt(2900000000)
	chainID := big.NewInt(1)
	dFeeTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     amount,
		Data:      nil,
	}
	txHex, txHash, _ := OfflineSignTx(dFeeTx, privateKeyHex, chainID)
	fmt.Println("txHex===", txHex, "txHash==", txHash)
}
