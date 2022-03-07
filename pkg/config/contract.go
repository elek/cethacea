package config

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/pkg/errors"
	"strings"
)

const DefaultContractFile = ".contracts.yaml"

type ContractRepo struct {
	Contracts  []*types.Contract
	Selected   string
	DefaultAbi string
}

func NewContractRepo(selected string, abi string, all bool) (*ContractRepo, error) {
	var contracts []*types.Contract

	err := LoadYamlConfig(DefaultContractFile, "contracts", &contracts)
	if err != nil {
		return nil, err
	}

	if len(contracts) == 1 && selected == "" {
		selected = contracts[0].Name
	}

	if selected != "" {

		_, err := hex.DecodeString(selected)
		if err == nil {
			contracts = append(contracts, &types.Contract{
				Name:    "<contract>",
				Address: "0x" + selected,
				Abi:     abi,
			})
			selected = "<contract>"
		}

		if strings.HasPrefix(selected, "0x") {
			_, err := hex.DecodeString(selected[2:])
			if err == nil {
				contracts = append(contracts, &types.Contract{
					Name:    "<contract>",
					Address: selected,
					Abi:     abi,
				})
				selected = "<contract>"
			}

		}
	}

	return &ContractRepo{
		Contracts:  contracts,
		Selected:   selected,
		DefaultAbi: abi,
	}, nil
}

func (c ContractRepo) ListContracts() ([]*types.Contract, error) {
	return c.Contracts, nil
}

func (c ContractRepo) AddContract(contract types.Contract) error {
	var contracts []*types.Contract
	err := LoadYamlConfig(DefaultContractFile, "contracts", &contracts)
	if err != nil {
		return err
	}

	updated := false
	for _, existing := range contracts {
		if existing.Name == contract.Name {
			existing.Address = contract.Address
			if contract.Abi != "" {
				existing.Abi = contract.Abi
			}
			if contract.ChainID != 0 {
				existing.ChainID = contract.ChainID
			}
			updated = true
			break
		}
	}

	if !updated {
		contracts = append(contracts, &contract)
	}
	return SaveYamlConfig(DefaultContractFile, &contracts)
}

func (c ContractRepo) GetContract(name string) (types.Contract, error) {
	for _, contract := range c.Contracts {
		if contract.Name == name {
			if c.DefaultAbi != "" {
				contract.Abi = c.DefaultAbi
			}
			return *contract, nil
		}
	}
	return types.Contract{}, errors.New(fmt.Sprintf("Contract '%s' is not found", name))
}

func (c ContractRepo) GetCurrentContract() (types.Contract, error) {
	return c.GetContract(c.Selected)
}
