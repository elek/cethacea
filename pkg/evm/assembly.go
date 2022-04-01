package evm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs/v2"
	"io/ioutil"
	"strings"
)

func asm(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	compiled, err := asmBytes(string(content))
	if err != nil {
		return err
	}

	dpl := deployable(compiled)

	encoded := hex.EncodeToString(compiled)
	fmt.Println(encoded)
	err = ioutil.WriteFile(file+".raw", []byte(encoded), 0644)
	if err != nil {
		return nil
	}
	err = ioutil.WriteFile(file+".bin", []byte(hex.EncodeToString(dpl)), 0644)
	if err != nil {
		return nil
	}
	return nil
}

type Compiler struct {
	jumpAddress map[string]int
	dataAddress map[string]int
	dataLength  map[string]int
}

func (c *Compiler) preProcess(code string) error {
	pc := 0
	for ix, line := range strings.Split(code, "\n") {
		line = strings.SplitN(line, ";", 2)[0]
		line = strings.TrimSpace(line)
		if line != "" {
			if strings.HasPrefix(line, ":") {
				c.jumpAddress[line[1:]] = pc
			} else if strings.HasPrefix(line, "#") {
				parts := strings.Split(line, " ")
				rawData, err := hexToBytes(parts[1])
				if err != nil {
					return err
				}
				c.dataAddress[parts[0][1:]] = pc
				c.dataLength[parts[0][1:]] = len(rawData)
				pc += len(rawData)
			} else {
				res, err := c.processLine(line, false)
				if err != nil {
					return errs.Errorf("Error in line %d %v", ix+1, err)
				}
				pc += len(res)
			}
		}
	}
	return nil
}

func (c *Compiler) compile(code string) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	for ix, line := range strings.Split(code, "\n") {
		line = strings.SplitN(line, ";", 2)[0]
		line = strings.TrimSpace(line)
		if line != "" {
			res, err := c.processLine(line, true)
			if err != nil {
				return nil, errs.Errorf("Error in line %d %v", ix+1, err)
			}
			buf.Write(res)
		}
	}
	return buf.Bytes(), nil
}

func NewCompiler() *Compiler {
	return &Compiler{
		jumpAddress: map[string]int{},
		dataLength:  map[string]int{},
		dataAddress: map[string]int{},
	}
}
func asmBytes(code string) ([]byte, error) {
	c := NewCompiler()
	err := c.preProcess(code)
	if err != nil {
		return nil, err
	}
	return c.compile(code)
}

func (c *Compiler) processLine(line string, resolve bool) ([]byte, error) {
	res := make([]byte, 0)
	if strings.HasPrefix(line, ":") {
		return res, nil
	}
	parts := strings.Split(line, " ")
	if strings.HasPrefix(line, "#") {
		rawData, err := hexToBytes(parts[1])
		if err != nil {
			return nil, err
		}
		return rawData, nil
	}
	opCodeString := parts[0]
	op := StringToOp(opCodeString)
	if op == 0 {
		return nil, fmt.Errorf("Unkown opcode %s", opCodeString)
	}
	res = append(res, byte(op))
	if op.IsPush() {
		if len(parts) != 2 {
			return nil, errs.Errorf("PUSH operation require parameters: " + line)
		}
		pushSize := int(byte(op)-0x60) + 1
		hexString := parts[1]

		if strings.HasPrefix(hexString, ":") || strings.HasPrefix(hexString, "#") {
			addressTable := c.jumpAddress
			symbol := hexString[1:]
			if strings.HasPrefix(hexString, "##") {
				addressTable = c.dataLength
				symbol = hexString[2:]
			} else if strings.HasPrefix(hexString, "#") {
				addressTable = c.dataAddress
			}

			if !resolve {
				res = append(res, make([]byte, pushSize)...)
			} else {
				position, found := addressTable[symbol]
				if !found {
					return nil, errs.Errorf("Instruction to non existent label: " + hexString)
				}
				intBytes, err := intInBytes(position, pushSize)
				if err != nil {
					return nil, errs.Errorf("Error in processing jump to %s %v", symbol, err)
				}
				res = append(res, intBytes...)

			}
		} else {
			hexString = strings.TrimPrefix(hexString, "0x")
			param, err := hex.DecodeString(hexString)
			if err != nil {
				return nil, err
			}

			if len(param) != pushSize {
				return nil, errs.Errorf(fmt.Sprintf("%s operation requires %d bytes long parameter", op.String(), pushSize))
			}
			res = append(res, param...)
		}
	}
	return res, nil
}

func intInBytes(position int, size int) ([]byte, error) {
	orig := position
	res := make([]byte, size)
	for i := size - 1; i >= 0; i-- {
		res[i] = byte(position % 256)
		position = position / 256
	}
	if position != 0 {
		return nil, errs.Errorf(fmt.Sprintf("jumping to %d position doesn't fit in byte %d", orig, size))
	}
	return res, nil
}

// deployable creates deployable code with simple code copy constructor.
func deployable(code []byte) []byte {
	deployCode := common.Hex2Bytes("630000000180600e6000396000f3")
	length := len(code)
	for i := 4; i >= 1; i-- {
		deployCode[i] = byte(length % 256)
		length = length / 256
	}
	return append(deployCode, code...)
}
