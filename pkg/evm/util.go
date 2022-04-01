package evm

import (
	"encoding/hex"
	"strings"
)

func hexToBytes(data string) ([]byte, error) {
	data = strings.TrimPrefix(data, "0x")
	return hex.DecodeString(data)
}

func paramSize(op OpCode) int {
	if op >= 0x60 && op <= 0x7f {
		return int(op) - 0x60 + 1
	}
	return 0
}
