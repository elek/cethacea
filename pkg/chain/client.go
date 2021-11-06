package chain

import (
	"context"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
)

type ChainClient interface {
	Balance(ctx context.Context, account common.Address) (decimal.Decimal, error)
	TokenBalance(ctx context.Context, token common.Address, account common.Address) (*big.Int, error)
	TokenInfo(ctx context.Context, token common.Address) (types.Item, error)
	GetTransaction(ctx context.Context, hash common.Hash) (types.Item, error)
	GetChainID(ctx context.Context) (int64, error)
	GetChainInfo(ctx context.Context) (types.Item, error)
	GetAccountInfo(ctx context.Context, account common.Address) (types.Item, error)
	SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error)
	SendQuery(ctx context.Context, from types.Account, to common.Address, options ...interface{}) ([]interface{}, error)
}

type WithData struct {
	Data []byte
}

type WithValue struct {
	Value *big.Int
}

type WithGasPrice struct {
	Price *big.Int
}

type WithGas struct {
	Gas uint64
}

type WithNonce struct {
	Nonce uint64
}

type WithGasFeeCap struct {
	Value *big.Int
}

type WithGasTipCap struct {
	Value *big.Int
}
