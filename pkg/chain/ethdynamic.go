package chain

import (
	"context"
	"encoding/hex"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

	tip, err := c.Client.SuggestGasTipCap(ctx)
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't get suggested gas price")
	}

	tx := ethtypes.DynamicFeeTx{
		To:        to,
		ChainID:   chainID,
		Nonce:     nonce,
		Gas:       8000000,
		GasTipCap: tip,
		GasFeeCap: new(big.Int).Mul(baseGas, big.NewInt(2)),
	}

	err = optionForDynamicTx(&tx, opts...)
	if err != nil {
		return hash, err
	}

	newTx := ethtypes.NewTx(&tx)
	signedTx, err := ethtypes.SignTx(newTx, ethtypes.NewLondonSigner(chainID), sender.PrivateKey())
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't sign the transaction")
	}
	log.Debug().
		Hex("from", sender.Address().Bytes()).
		Str("to", optionalAddress(signedTx.To())).
		Str("value", signedTx.Value().String()).
		Str("data", hex.EncodeToString(signedTx.Data())).
		Msg("Sending transaction")

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
		case WithGas:
			tx.Gas = o.Gas
		default:
			return errors.Errorf("Unsupported option type %t:", opt)
		}
	}
	return nil
}
