package chain

import (
	"context"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

func (c *Eth) sendRawLegacyTx(ctx context.Context, sender types.Account, to *common.Address, opts ...interface{}) (hash common.Hash, err error) {

	nonce, err := c.Client.PendingNonceAt(ctx, sender.Address())
	if err != nil {
		return hash, err
	}

	chainId, err := c.getChainId(ctx)
	if err != nil {
		return hash, err
	}

	gasPrice, err := c.Client.SuggestGasPrice(ctx)
	if err != nil {
		return hash, err
	}

	tx := ethtypes.LegacyTx{
		To:       to,
		Nonce:    nonce,
		Gas:      8000000,
		GasPrice: gasPrice,
	}

	err = optionForLegacyTx(&tx, opts...)
	if err != nil {
		return hash, err
	}

	newTx := ethtypes.NewTx(&tx)
	signedTx, err := ethtypes.SignTx(newTx, ethtypes.NewLondonSigner(chainId), sender.PrivateKey())
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't sign the transaction")
	}

	err = c.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return hash, errors.Wrap(err, "Couldn't send transaction")
	}

	return signedTx.Hash(), nil
}

func optionForLegacyTx(tx *ethtypes.LegacyTx, opts ...interface{}) error {
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
