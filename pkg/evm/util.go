package evm

import (
	"encoding/hex"
	"strings"
)

func hexToBytes(data string) ([]byte, error) {
	if strings.HasPrefix(data, "0x") {
		data = data[2:]
	}
	return hex.DecodeString(data)
}

func paramSize(op OpCode) int {
	if op >= 0x60 && op <= 0x7f {
		return int(op) - 0x60 + 1
	}
	return 0
}
