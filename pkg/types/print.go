package types

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
)

func PrettyETH(wei *big.Int) string {
	switch {
	case wei.Cmp(big.NewInt(1_000_000_000_000_000)) > 0:
		return fmt.Sprintf("%s ETH", decimal.NewFromBigInt(wei, -18))
	case wei.Cmp(big.NewInt(1_000_000_0)) > 0:
		return fmt.Sprintf("%s GWei", decimal.NewFromBigInt(wei, -9))
	default:
		return fmt.Sprintf("%s Wei", wei)
	}
}
