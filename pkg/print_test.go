package cethacea

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrintEthFromDecimal(t *testing.T) {
	d := decimal.New(1, 0)
	require.Equal(t, "1000000000000000000 (1 ETH)", PrintEthFromDecimal(d))
}
