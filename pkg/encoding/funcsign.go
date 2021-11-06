package encoding

import (
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

type FunctionSignature abi.Method

func (f *FunctionSignature) CanonicalName() string {
	argStrings := make([]string, 0)
	for _, x := range f.Inputs {
		argStrings = append(argStrings, x.Type.String())
	}
	return fmt.Sprintf("%s(%s)", f.RawName, strings.Join(argStrings, ","))
}

func (f *FunctionSignature) EncodeFuncCall(resolver types.AddressResolver, args ...string) ([]byte, error) {
	if len(f.Inputs) != len(args) {
		return nil, fmt.Errorf("argument types (%d) and values (%d), has different length", len(f.Inputs), len(args))
	}
	bytes, err := EncodeArguments(resolver, f.Inputs, args)
	if err != nil {
		return nil, err
	}
	c := bytes
	if f.Name != "" {
		c = append(FunctionHash(f.CanonicalName()), bytes...)
	}
	return c, nil
}

func ParseFunctionSignature(sign string) (FunctionSignature, error) {
	nameArg := strings.Split(sign, "(")
	paramParts := strings.Split(nameArg[1], ")")
	argumentsString := paramParts[0]
	argTypes := strings.Split(argumentsString, ",")
	if argumentsString == "" {
		argTypes = []string{}
	}
	returnTypes := strings.Split(paramParts[1], ",")
	if paramParts[1] == "" {
		returnTypes = []string{}
	}

	inputs := abi.Arguments{}
	for _, argType := range argTypes {
		ty, err := abi.NewType(argType, "", nil)
		if err != nil {
			return FunctionSignature{}, err
		}
		inputs = append(inputs, abi.Argument{
			Type: ty,
		})
	}

	outputs := abi.Arguments{}
	for _, returnType := range returnTypes {
		ty, err := abi.NewType(returnType, "", nil)
		if err != nil {
			return FunctionSignature{}, err
		}
		outputs = append(outputs, abi.Argument{
			Type: ty,
		})
	}

	return FunctionSignature{
		Name:    nameArg[0],
		RawName: nameArg[0],
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}
