package chain

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

type Eth struct {
	Client    *ethclient.Client
	legacy    bool
	noop      bool
	confirm   bool
	gas       uint64
	gasTipCap *big.Int
}

func (c *Eth) SendQuery(ctx context.Context, from common.Address, to common.Address, options ...interface{}) ([]byte, error) {
	return SendQuery(ctx, c.Client, from, to, options...)
}

func (c *Eth) GetAccountInfo(ctx context.Context, account common.Address) (types.Item, error) {
	i := types.Item{
		Record: types.Record{
			Fields: []types.Field{
				{
					Name:  "address",
					Value: account.String(),
				},
			},
		},
	}
	return i, nil
}

var _ ChainClient = &Eth{}

func NewEth(url string, confirm bool, gas uint64, gasTipCap *big.Int) (*Eth, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create ethereum client with url %s", url)
	}
	return &Eth{
		confirm:   confirm,
		Client:    client,
		gas:       gas,
		gasTipCap: gasTipCap,
	}, nil

}

func (c *Eth) getChainId(ctx context.Context) (*big.Int, error) {
	return c.Client.ChainID(ctx)
}

func (c *Eth) Query(ctx context.Context, resolver types.AddressResolver, sender common.Address, contract common.Address, function string, args ...string) ([]interface{}, error) {
	return Query(ctx, c.Client, resolver, sender, contract, function, args...)
}

func (c *Eth) SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error) {
	return c.sendRawTransaction(ctx, from, to, options...)
}

func (c *Eth) Call(ctx context.Context, sender types.Account, contract common.Address, function string, argTypes abi.Arguments, args ...interface{}) (*ethtypes.Receipt, error) {
	if c.confirm {
		fmt.Printf("function:      %s\n", function)
		for i, a := range argTypes {
			fmt.Printf("   %s: %s\n", a.Name, args[i])
		}
	}
	data, err := FunctionCallData(function, argTypes, args)
	if err != nil {
		return nil, err
	}

	txHash, err := c.sendRawTransaction(ctx, sender, &contract, WithData{data})
	if err != nil {
		return nil, errors.Wrap(err, "CallContract is failed")
	}

	var receipt *ethtypes.Receipt
	for {
		receipt, err = c.Client.TransactionReceipt(ctx, txHash)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break

	}
	return receipt, nil
}

func (c *Eth) Balance(ctx context.Context, account common.Address) (decimal.Decimal, error) {
	at, err := c.Client.BalanceAt(ctx, account, nil)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return decimal.NewFromBigInt(at, -18), err
}

func (c *Eth) TokenBalance(ctx context.Context, token common.Address, account common.Address) (*big.Int, error) {
	res, err := c.Query(ctx, types.WithoutAddressResolution{}, account, token, "balanceOf(address)uint256", account.String())
	if err != nil {
		return big.NewInt(0), err
	}

	return res[0].(*big.Int), nil
}

func (c *Eth) TokenInfo(ctx context.Context, token common.Address) (TokenInfo, error) {
	t := TokenInfo{}
	symbol, err := c.Query(ctx, types.WithoutAddressResolution{}, token, token, "symbol()string")
	if err == nil {
		t.Symbol = symbol[0].(string)
	}
	dec, err := c.Query(ctx, types.WithoutAddressResolution{}, token, token, "decimals()uint8")
	if err == nil {
		t.Decimal = dec[0].(uint8)
	}
	t.Address = token
	return t, nil
}

func (c *Eth) GetTransaction(ctx context.Context, hash common.Hash) (types.Item, error) {
	return GetTransaction(ctx, c.Client, hash)
}

func optionalAddress(to *common.Address) string {
	if to == nil {
		return "<nil>"
	}
	return to.Hex()
}

func (c *Eth) GetChainID(ctx context.Context) (int64, error) {
	chainId, err := c.Client.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return chainId.Int64(), err
}

func (c *Eth) GetChainInfo(ctx context.Context) (types.Item, error) {
	r := types.Record{
		Fields: []types.Field{},
	}

	chainID, err := c.Client.ChainID(ctx)
	if err != nil {
		r.AddField("chainID", "??? "+err.Error())
	} else {
		r.AddField("chainID", chainID)
	}

	networkID, err := c.Client.NetworkID(ctx)
	if err != nil {
		r.AddField("networkID", "??? "+err.Error())
	} else {
		r.AddField("networkID", networkID)
	}

	gasTipCap, err := c.Client.SuggestGasTipCap(ctx)
	if err != nil {
		r.AddField("gasTipCap", "??? "+err.Error())
	} else {
		r.AddField("gasTipCap", gasTipCap)
	}

	lastBlock, err := c.Client.BlockByNumber(ctx, nil)
	if err != nil {
		r.AddField("lastGasFee", "??? "+err.Error())
	} else {
		r.AddField("lastGasFee", lastBlock.BaseFee())
	}

	return types.Item{
		Record: r,
	}, nil
}
