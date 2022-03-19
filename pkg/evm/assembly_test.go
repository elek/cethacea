package evm

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawData(t *testing.T) {
	code := `
PUSH4 #data
#data 0x0000EEFF
`
	out, err := asmBytes(code)
	require.Nil(t, err)
	require.Equal(t, "63000000050000eeff", hex.EncodeToString(out))
}

func TestRawDataLength(t *testing.T) {
	code := `
PUSH4 ##data
#data 0x0000EEFF
`
	out, err := asmBytes(code)
	require.Nil(t, err)
	require.Equal(t, "63000000040000eeff", hex.EncodeToString(out))
}
