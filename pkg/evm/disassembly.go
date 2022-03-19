package evm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
)

func disasm(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	//rawContent, err := base58.Decode(string(content))
	rawContent := common.Hex2Bytes(string(content))
	if err != nil {
		return err
	}

	for c := 0; c < len(rawContent); c++ {
		op := OpCode(rawContent[c])
		pc := paramSize(op)
		params := make([]byte, pc)
		for p := 0; p < pc; p++ {
			c++
			params[p] = rawContent[c]
		}
		if len(params) > 0 {
			fmt.Printf("%d %s %v\n", c, op, params)
		} else {
			fmt.Printf("%d %s\n", c, op)
		}
	}

	return nil
}
