package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/zeebo/errs/v2"
	"github.com/zksync-sdk/zksync2-go"
	"math/big"
)

type Zksync2 struct {
	zk *zksync2.DefaultProvider
}

var _ ChainClient = &Zksync2{}

func NewZksync2(url string, confirm bool) (*Zksync2, error) {
	zkSyncProvider, err := zksync2.NewDefaultProvider(url)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return &Zksync2{
		zk: zkSyncProvider,
	}, nil
}

func (z *Zksync2) Balance(ctx context.Context, account common.Address) (decimal.Decimal, error) {
	val, err := z.zk.GetBalance(account, zksync2.BlockNumberCommitted)
	if err != nil {
		return decimal.Decimal{}, nil
	}
	return decimal.NewFromBigInt(val, -18), nil
}

func (z *Zksync2) TokenBalance(ctx context.Context, token common.Address, account common.Address) (*big.Int, error) {
	res, err := Query(ctx, z.zk.Client, types.WithoutAddressResolution{}, account, token, "balanceOf(address)uint256", account.String())
	if err != nil {
		return big.NewInt(0), err
	}

	return res[0].(*big.Int), nil
}

func (z *Zksync2) TokenInfo(ctx context.Context, token common.Address) (TokenInfo, error) {
	t := TokenInfo{}
	symbol, err := Query(ctx, z.zk.Client, types.WithoutAddressResolution{}, token, token, "symbol()string")
	if err == nil {
		t.Symbol = symbol[0].(string)
	}
	dec, err := Query(ctx, z.zk.Client, types.WithoutAddressResolution{}, token, token, "decimals()uint8")
	if err == nil {
		t.Decimal = dec[0].(uint8)
	}
	t.Address = token
	return t, nil
}

func (z *Zksync2) GetTransaction(ctx context.Context, hash common.Hash) (types.Item, error) {
	return GetTransaction(ctx, z.zk.Client, hash)
}

func (z *Zksync2) GetChainID(ctx context.Context) (int64, error) {
	v, err := z.zk.ChainID(ctx)
	return v.Int64(), err
}

func (z *Zksync2) GetChainInfo(ctx context.Context) (types.Item, error) {
	r := types.Record{
		Fields: []types.Field{},
	}

	chainID, err := z.zk.ChainID(ctx)
	if err != nil {
		r.AddField("chainID", "??? "+err.Error())
	} else {
		r.AddField("chainID", chainID)
	}

	networkID, err := z.zk.NetworkID(ctx)
	if err != nil {
		r.AddField("networkID", "??? "+err.Error())
	} else {
		r.AddField("networkID", networkID)
	}

	bc, err := z.zk.ZksGetBridgeContracts()
	if err != nil {
		r.AddField("bridge", "??? "+err.Error())
	} else {
		r.AddField("l1Erc20Bridge", bc.L1Erc20DefaultBridge)
		r.AddField("l2Erc20Bridge", bc.L2Erc20DefaultBridge)
		r.AddField("l1ETHBridge", bc.L1EthDefaultBridge)
		r.AddField("l2ETHBridge", bc.L2EthDefaultBridge)
	}

	return types.Item{
		Record: r,
	}, nil
}

func (z *Zksync2) GetAccountInfo(ctx context.Context, account common.Address) (types.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (z *Zksync2) SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error) {
	tx := zksync2.CreateFunctionCallTransaction(
		from.Address(),
		*to,
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		hexutil.Bytes{},
		nil, nil,
	)
	err := optionForZksyncTx(tx, options...)
	if err != nil {
		return common.Hash{}, errs.Wrap(err)
	}

	nonce, err := z.zk.NonceAt(ctx, from.Address(), nil)
	if err != nil {
		return common.Hash{}, errs.Wrap(err)
	}

	gas, err := z.zk.EstimateGas(tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to EstimateGas: %w", err)
	}

	gasPrice, err := z.zk.GetGasPrice()
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to GetGasPrice: %w", err)
	}

	chainId, err := z.zk.ChainID(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to GetChainID: %w", err)
	}

	data := zksync2.NewTransaction712(
		chainId,
		big.NewInt(int64(nonce)),
		gas,
		*to,
		big.NewInt(0), // value
		tx.Data,
		big.NewInt(100000000), // TODO: Estimate correct one
		gasPrice,
		from.Address(),
		&zksync2.Eip712Meta{
			ErgsPerPubdata: zksync2.NewBig(160000),
		})

	ethereumSigner, err := zksync2.NewEthSignerFromRawPrivateKey(from.PrivateKey().D.Bytes(), 280)
	if err != nil {
		return common.Hash{}, err
	}

	domain := ethereumSigner.GetDomain()
	signature, err := ethereumSigner.SignTypedData(domain, data)
	if err != nil {
		return common.Hash{}, err
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			data.GetEIP712Type():   data.GetEIP712Types(),
			domain.GetEIP712Type(): domain.GetEIP712Types(),
		},
		PrimaryType: data.GetEIP712Type(),
		Domain:      domain.GetEIP712Domain(),
		Message:     data.GetEIP712Message(),
	}
	hash, err := ethereumSigner.HashTypedData(typedData)
	if err != nil {
		return common.Hash{}, err
	}
	fmt.Println(hex.EncodeToString(hash))
	h := common.Hash{}
	copy(h[:], hash)
	rawTx, err := data.RLPValues(signature)
	_, err = z.zk.SendRawTransaction(rawTx)
	if err != nil {
		return h, err
	}
	fmt.Println(hex.EncodeToString(crypto.Keccak256(rawTx)))

	return h, nil

}

func (z *Zksync2) SendQuery(ctx context.Context, from common.Address, to common.Address, options ...interface{}) ([]byte, error) {
	return SendQuery(ctx, z.zk.Client, from, to, options)
}

func optionForZksyncTx(tx *zksync2.Transaction, opts ...interface{}) error {
	for _, opt := range opts {
		switch o := opt.(type) {
		case WithData:
			tx.Data = o.Data
		case WithValue:
			val := hexutil.Big(*o.Value)
			tx.Value = &val
		default:
			return errors.Errorf("Unsupported option type %t:", opt)
		}
	}
	return nil
}
