package types

import (
	"encoding/json"
	"github.com/elek/cethacea/pkg/standards"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Contract struct {
	Name    string
	Address string
	Abi     string
	Type    string
	ChainID int64
}

func (c Contract) GetAbi() (parsedAbi abi.ABI, err error) {
	abiContent, found := standards.GetPredefinedContract(c.Abi)
	if !found {
		abiContent, err = ioutil.ReadFile(c.Abi)
		if err != nil {
			return abi.ABI{}, errors.Wrap(err, "couldn't open ABI file "+c.Abi)
		}
	}

	err = json.Unmarshal(abiContent, &parsedAbi)
	if err == nil {
		return parsedAbi, nil
	}

	wrapped := struct {
		Abi abi.ABI
	}{}
	err = json.Unmarshal(abiContent, &wrapped)
	if err != nil {
		return abi.ABI{}, errors.Wrap(err, "ABI file is not a valid JSON: "+c.Abi)
	}
	return wrapped.Abi, nil
}

func (c Contract) GetAddress() common.Address {
	return common.HexToAddress(c.Address)
}
