package types

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type AddressResolver interface {
	ResolveAddress(address string) (common.Address, error)
}

type WithoutAddressResolution struct {
}

func (WithoutAddressResolution) ResolveAddress(address string) (common.Address, error) {
	return common.HexToAddress(address), nil
}

type Balance interface {
	Balance(ctx context.Context, account common.Address) (*big.Int, error)
	TokenBalance(ctx context.Context, account common.Address, token common.Address) (*big.Int, error)
}
