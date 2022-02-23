package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/shopspring/decimal"

	"github.com/elek/cethacea/pkg/encoding"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type Eth struct {
	Client *ethclient.Client
	legacy bool
}

func (c *Eth) GetAccountInfo(ctx context.Context, account common.Address) (types.Item, error) {
	panic("implement me")
}

func (c *Eth) SendQuery(ctx context.Context, from types.Account, to common.Address, options ...interface{}) ([]interface{}, error) {
	panic("implement me")
}

var _ ChainClient = &Eth{}

func NewEthFromURL(url string) (*Eth, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create ethereum client with url %s", url)
	}
	return &Eth{
		Client: client,
	}, nil

}

func (c *Eth) getChainId(ctx context.Context) (*big.Int, error) {
	return c.Client.ChainID(ctx)
}

func (c *Eth) Query(ctx context.Context, resolver types.AddressResolver, sender common.Address, contract common.Address, function string, args ...string) ([]interface{}, error) {
	fs, err := encoding.ParseFunctionSignature(function)
	if err != nil {
		return nil, err
	}
	data, err := fs.EncodeFuncCall(resolver, args...)
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{
		From: sender,
		To:   &contract,
		Data: data,
	}
	res, err := c.Client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "CallContract is failed")
	}

	values, err := fs.Outputs.Unpack(res)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (c *Eth) SubmitTx(ctx context.Context, sender types.Account, to *common.Address, opts ...interface{}) (common.Hash, error) {
	if !c.legacy {
		return c.sendRawTransaction(ctx, sender, to, opts...)
	} else {
		return c.sendRawLegacyTx(ctx, sender, to, opts...)
	}
}

func (c *Eth) Call(ctx context.Context, sender types.Account, contract common.Address, function string, argTypes abi.Arguments, args ...interface{}) (*ethtypes.Receipt, error) {
	data, err := c.FunctionCallData(function, argTypes, args)
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

func (c *Eth) FunctionCallData(function string, argTypes abi.Arguments, args ...interface{}) ([]byte, error) {
	data, err := argTypes.Pack(args...)
	if err != nil {
		return nil, errors.Wrap(err, "Arguments couldn't be packed")
	}
	data = append(encoding.FunctionHash(function), data...)
	return data, nil
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

func (c *Eth) TokenInfo(ctx context.Context, token common.Address) (types.Item, error) {
	r := types.Item{
		Record: types.Record{
			Fields: []types.Field{},
		},
	}

	res, err := c.Query(ctx, types.WithoutAddressResolution{}, token, token, "symbol()string")
	if err == nil {
		r.Record.AddField("symbol", res[0])
	}
	res, err = c.Query(ctx, types.WithoutAddressResolution{}, token, token, "decimals()uint8")
	if err == nil {
		r.Record.AddField("decimals", res[0])
	}
	return r, nil
}
func (c *Eth) SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error) {
	return c.sendRawTransaction(ctx, from, to, options...)
}

func (c *Eth) GetTransaction(ctx context.Context, hash common.Hash) (types.Item, error) {
	tx, _, err := c.Client.TransactionByHash(ctx, hash)
	if err != nil {
		return types.Item{}, errors.Wrap(err, "Couldn't read transaction")
	}

	i := types.Item{
		Record: types.Record{
			Fields: []types.Field{
				{
					Name:  "hash",
					Value: tx.Hash(),
				},

				{
					Name:  "to",
					Value: optionalAddress(tx.To()),
				},
				{
					Name:  "value",
					Value: tx.Value(),
				},
				{
					Name:  "gas",
					Value: tx.Gas(),
				},
				{
					Name:  "gasFeeCap",
					Value: tx.GasFeeCap(),
				},
				{
					Name:  "gasTipCap",
					Value: tx.GasTipCap(),
				},
				{
					Name:  "type",
					Value: tx.Type(),
				},
				{
					Name:  "data",
					Value: hex.EncodeToString(tx.Data()),
				},
			},
		},
	}

	receipt, err := c.Client.TransactionReceipt(ctx, hash)
	if err != nil {
		fmt.Println("Couldn't read transaction receipt: " + err.Error())
	} else {
		i.Record.Fields = append(i.Record.Fields, []types.Field{
			{
				Name:  "status",
				Value: receipt.Status,
			},
			{
				Name:  "block",
				Value: receipt.BlockHash.String(),
			},
			{
				Name:  "contract",
				Value: receipt.ContractAddress.String(),
			},
			{
				Name:  "gasUsed",
				Value: receipt.GasUsed,
			},
			{
				Name:  "cumulative",
				Value: receipt.CumulativeGasUsed,
			},
		}...)
	}

	return i, nil
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
