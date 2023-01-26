package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/encoding"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func (c *Eth) SendQuery(ctx context.Context, sender common.Address, to common.Address, options ...interface{}) ([]byte, error) {

	data := []byte{}

	for _, option := range options {
		switch o := option.(type) {
		case WithData:
			data = o.Data
		default:
			return nil, errors.Errorf("unsupported option %v", o)
		}
	}
	msg := ethereum.CallMsg{
		From: sender,
		To:   &to,
		Data: data,
	}

	res, err := c.Client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Debug().
			Err(err).
			Hex("data", data).
			Hex("to", to.Bytes()).
			Hex("from", sender.Bytes()).
			Msg("CallContract")
		return nil, errors.Wrap(err, "CallContract is failed")
	}

	log.Debug().
		Hex("data", data).
		Hex("to", to.Bytes()).
		Hex("from", sender.Bytes()).
		Hex("response", res).
		Msg("CallContract")

	return res, nil
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
	fs, err := encoding.ParseFunctionSignature(function)
	if err != nil {
		return nil, err
	}
	data, err := fs.EncodeFuncCall(resolver, args...)
	if err != nil {
		return nil, err
	}
	res, err := c.SendQuery(ctx, sender, contract, WithData{Data: data})
	if err != nil {
		return nil, err
	}
	values, err := fs.Outputs.Unpack(res)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (c *Eth) SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error) {
	return c.sendRawTransaction(ctx, from, to, options...)
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
	if c.confirm {
		fmt.Printf("function:      %s\n", function)
		for i, a := range argTypes {
			fmt.Printf("   %s: %s\n", a.Name, args[i])
		}
	}
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
					Name:    "gasFeeCap",
					Value:   tx.GasFeeCap(),
					Printer: types.EthPrintType,
				},
				{
					Name:    "gasTipCap",
					Value:   tx.GasTipCap(),
					Printer: types.EthPrintType,
				},
				{
					Name:  "type",
					Value: tx.Type(),
				},
				{
					Name:  "nonce",
					Value: tx.Nonce(),
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
		return i, err
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

		block, err := c.Client.BlockByHash(ctx, receipt.BlockHash)
		if err != nil {
			return i, err
		}
		i.Record.Fields = append(i.Record.Fields,
			types.Field{
				Name:    "base-gas-fee",
				Value:   block.BaseFee(),
				Printer: types.EthPrintType,
			},
		)
		i.Record.Fields = append(i.Record.Fields,
			types.Field{
				Name:    "fee",
				Value:   new(big.Int).Mul(block.BaseFee(), big.NewInt(int64(receipt.GasUsed))),
				Printer: types.EthPrintType,
			},
		)
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
