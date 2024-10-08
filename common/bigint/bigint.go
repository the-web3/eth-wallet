package bigint

import "math/big"

var (
	Zero = big.NewInt(0)
	One  = big.NewInt(1)
)

func Clamp(start, end *big.Int, size uint64) *big.Int {
	temp := new(big.Int)
	count := temp.Sub(end, start).Uint64() + 1
	if count <= size {
		return end
	}

	temp.Add(start, big.NewInt(int64(size-1)))
	return temp
}

func Matcher(num int64) func(*big.Int) bool {
	return func(bi *big.Int) bool { return bi.Int64() == num }
}

func WeiToETH(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetString(wei.String())
	return f.Quo(f, big.NewFloat(1e18))
}
