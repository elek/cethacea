package encoding

import (
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strconv"
	"strings"
)

func FunctionHash(f string) []byte {
	return crypto.Keccak256Hash([]byte(f)).Bytes()[0:4]
}

func EncodeArguments(resolver types.AddressResolver, arguments abi.Arguments, args []string) (res []byte, err error) {
	typedValues := make([]interface{}, 0)
	for ix, arg := range args {
		switch arguments[ix].Type.T {
		case abi.StringTy:
			typedValues = append(typedValues, arg)
		case abi.AddressTy:
			addr, err := resolver.ResolveAddress(arg)
			if err != nil {
				return nil, err
			}
			typedValues = append(typedValues, addr)
		case abi.UintTy, abi.IntTy:
			var bi = new(big.Int)
			if strings.HasPrefix(arg, "0x") {
				bi.SetString(arg[2:], 16)
			} else {
				var ok bool
				bi, ok = new(big.Int).SetString(arg, 10)
				if !ok {
					return nil, fmt.Errorf("wrong big int format")
				}
			}
			typedValues = append(typedValues, &bi)
		case abi.BytesTy:
			val, err := HexToBytes(arg)
			if err != nil {
				return nil, err
			}
			size := arguments[ix].Type.Size
			if size == 0 {
				size = len(val)
			}
			c := make([]byte, size)
			copy(c[:], val)
			typedValues = append(typedValues, c)
		case abi.FixedBytesTy:
			val, err := HexToBytes(arg)
			if err != nil {
				return nil, err
			}
			size := arguments[ix].Type.Size
			typedValues = append(typedValues, SliceToArray(size, val))
		case abi.ArrayTy:

			val, err := HexToBytes(arg)
			if err != nil {
				return nil, err
			}
			size := arguments[ix].Type.Size
			typedValues = append(typedValues, SliceToArray(size, val))
		case abi.BoolTy:
			val := false
			for _, t := range []string{"1", "true"} {
				if strings.ToLower(arg) == t {
					val = true
				}
			}
			typedValues = append(typedValues, val)
		default:
			return nil, fmt.Errorf("unsupported argument type %v", arguments[ix])
		}
	}

	bytes, err := arguments.Pack(
		typedValues...,
	)
	return bytes, err
}

func SliceToArray(size int, val []byte) interface{} {
	var c interface{}
	switch size {
	case 32:
		arr := [32]byte{}
		copy(arr[:], val)
		c = arr
	case 16:
		arr := [16]byte{}
		copy(arr[:], val)
		c = arr
	case 8:
		arr := [8]byte{}
		copy(arr[:], val)
		c = arr
	default:
		panic("Unsupported array size " + strconv.Itoa(size))
	}

	return c
}

func HexToBytes(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, "0x")
	return hex.DecodeString(data)
}
