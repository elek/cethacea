package chain

import (
	"context"
	"encoding/hex"
	"github.com/elek/cethacea/pkg/encoding"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"math/big"
)

func Query(ctx context.Context, client *ethclient.Client, resolver types.AddressResolver, sender common.Address, contract common.Address, function string, args ...string) ([]interface{}, error) {
	fs, err := encoding.ParseFunctionSignature(function)
	if err != nil {
		return nil, err
	}
	data, err := fs.EncodeFuncCall(resolver, args...)
	if err != nil {
		return nil, err
	}
	res, err := SendQuery(ctx, client, sender, contract, WithData{Data: data})
	if err != nil {
		return nil, err
	}
	values, err := fs.Outputs.Unpack(res)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func SendQuery(ctx context.Context, client *ethclient.Client, sender common.Address, to common.Address, options ...interface{}) ([]byte, error) {
	data := []byte{}

	for _, option := range options {
		switch o := option.(type) {
		case WithData:
			data = o.Data
		default:
			return nil, errors.Errorf("unsupported option %T", o)
		}
	}
	msg := ethereum.CallMsg{
		From: sender,
		To:   &to,
		Data: data,
	}

	res, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Debug().
			Err(err).
			Hex("data", data).
			Hex("to", to.Bytes()).
			Hex("from", sender.Bytes()).
			Hex("resp", res).
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

func FunctionCallData(function string, argTypes abi.Arguments, args ...interface{}) ([]byte, error) {
	data, err := argTypes.Pack(args...)
	if err != nil {
		return nil, errors.Wrap(err, "Arguments couldn't be packed")
	}
	data = append(encoding.FunctionHash(function), data...)
	return data, nil
}

func GetTransaction(ctx context.Context, client *ethclient.Client, hash common.Hash) (types.Item, error) {
	tx, _, err := client.TransactionByHash(ctx, hash)
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

	receipt, err := client.TransactionReceipt(ctx, hash)
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

		block, err := client.BlockByHash(ctx, receipt.BlockHash)
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
