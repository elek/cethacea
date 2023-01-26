package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/zeebo/errs/v2"
	"math/big"
)

func (c *Eth) sendRawTransaction(ctx context.Context, sender types.Account, to *common.Address, opts ...interface{}) (hash common.Hash, err error) {
	nonce, err := c.Client.PendingNonceAt(ctx, sender.Address())
	if err != nil {
		return hash, err
	}

	chainID, err := c.getChainId(ctx)
	if err != nil {
		return hash, err
	}

	baseGas, err := c.Client.SuggestGasPrice(ctx)
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't get suggested gas price")
	}

	tip := c.gasTipCap
	if tip == nil {
		tip, err = c.Client.SuggestGasTipCap(ctx)
		if err != nil {
			return hash, errors.Wrap(err, "Couldn't get suggested gas price")
		}
	}

	tx := ethtypes.DynamicFeeTx{
		To:        to,
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tip,
	}

	err = optionForDynamicTx(&tx, opts...)
	if err != nil {
		return hash, err
	}

	gas := c.gas
	if gas == 0 {
		gas, err = c.Client.EstimateGas(context.Background(), ethereum.CallMsg{
			From: sender.Address(),
			To:   to,
			Data: tx.Data,
		})
		if err != nil {
			return hash, err
		}
		gas = gas * 13 / 10
	}
	tx.Gas = uint64(gas)

	tx.GasFeeCap = new(big.Int).Add(new(big.Int).Mul(baseGas, big.NewInt(2)), tx.GasTipCap)

	newTx := ethtypes.NewTx(&tx)
	signedTx, err := ethtypes.SignTx(newTx, ethtypes.NewLondonSigner(chainID), sender.PrivateKey())
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't sign the transaction")
	}

	log.Debug().Hex("from", sender.Address().Bytes()).
		Str("to", optionalAddress(signedTx.To())).
		Str("value", signedTx.Value().String()).
		Str("data", hex.EncodeToString(signedTx.Data())).
		Int64("gasTipCap", signedTx.GasTipCap().Int64()).
		Int64("gasFeeCap", signedTx.GasFeeCap().Int64()).
		Msg("eth_sendRawTransaction")

	if c.confirm {
		fmt.Printf("from:          %s\n", sender.Address().String())
		fmt.Printf("to:            %s\n", optionalAddress(signedTx.To()))
		fmt.Printf("value:         %s\n", types.PrettyETH(signedTx.Value()))
		fmt.Printf("data:          %x\n", signedTx.Data())
		fmt.Printf("gas-tip-cap:   %s\n", types.PrettyETH(signedTx.GasTipCap()))
		fmt.Printf("gas-fee-cap:   %s\n", types.PrettyETH(signedTx.GasFeeCap()))
		fmt.Printf("gas:           %d\n", signedTx.Gas())
		fmt.Printf("max-gas-price: %s\n", types.PrettyETH(new(big.Int).Mul(signedTx.GasFeeCap(), big.NewInt(int64(signedTx.Gas())))))
		fmt.Println("Are you sure to send it?")
		s := ""
		_, err := fmt.Scanln(&s)
		if err != nil {
			return common.Hash{}, errs.Wrap(err)
		}
		if s != "y" {
			panic("not confirmed")
		}

	}
	err = c.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't send transaction")
	}

	return signedTx.Hash(), nil
}

func optionForDynamicTx(tx *ethtypes.DynamicFeeTx, opts ...interface{}) error {
	for _, opt := range opts {
		switch o := opt.(type) {
		case WithData:
			tx.Data = o.Data
		case WithValue:
			tx.Value = o.Value
		case WithNonce:
			tx.Nonce = o.Nonce
		case WithGasTipCap:
			tx.GasTipCap = o.Value
		case WithGasFeeCap:
			tx.GasFeeCap = o.Value
		case WithGas:
			tx.Gas = o.Gas
		default:
			return errors.Errorf("Unsupported option type %t:", opt)
		}
	}
	return nil
}
