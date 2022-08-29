package chain

import (
	"context"
	"encoding/json"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"net/http"
)

type AccountInfo struct {
	ID        int64
	Address   string
	Committed AccountState
	Verified  AccountState
}
type AccountState struct {
	Balances map[string]string
	Nonce    uint64
}

type ZkSync struct {
	URL string
}

func (c *ZkSync) GetAccountInfo(ctx context.Context, account common.Address) (types.Item, error) {
	item := types.NewItem()

	client, err := rpc.DialContext(ctx, c.URL+"jsrpc")
	if err != nil {
		return item, err
	}
	defer client.Close()

	result := AccountInfo{}
	err = client.CallContext(ctx, &result, "account_info", account.String())
	if err != nil {
		return item, err
	}
	balanceString := result.Committed.Balances["ETH"]
	balance := big.NewInt(0)
	if balanceString != "" {
		balance, _ = new(big.Int).SetString(balanceString, 10)
	}
	item.AddField("balance", decimal.NewFromBigInt(balance, -18))
	item.AddField("id", result.ID)
	return item, nil
}

var _ ChainClient = &ZkSync{}

func NewZkSyncFromURL(url string) (*ZkSync, error) {
	if url[len(url)-1] != '/' {
		url = url + "/"
	}
	return &ZkSync{
		URL: url,
	}, nil
}

func (c *ZkSync) Balance(ctx context.Context, account common.Address) (decimal.Decimal, error) {
	client, err := rpc.DialContext(ctx, c.URL+"jsrpc")
	if err != nil {
		return decimal.Decimal{}, err
	}
	defer client.Close()
	result := AccountInfo{}
	err = client.CallContext(ctx, &result, "account_info", account.String())
	if err != nil {
		return decimal.Decimal{}, err
	}
	balanceString := result.Committed.Balances["ETH"]
	balance := big.NewInt(0)
	if balanceString != "" {
		balance, _ = new(big.Int).SetString(balanceString, 10)
	}
	return decimal.NewFromBigInt(balance, -18), nil
}

func (c *ZkSync) TokenBalance(ctx context.Context, account common.Address, token common.Address) (*big.Int, error) {
	panic("not yet implemented")
}

func (c *ZkSync) GetTransaction(ctx context.Context, hash common.Hash) (types.Item, error) {
	panic("implement me")
}

func (c *ZkSync) GetChainID(ctx context.Context) (int64, error) {
	panic("implement me")
}

func (c *ZkSync) GetChainInfo(ctx context.Context) (types.Item, error) {
	r := types.Record{
		Fields: []types.Field{},
	}
	url := c.URL + "api/v0.2/config"
	resp, err := http.Get(url)
	if err != nil {
		return types.Item{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return types.Item{}, err
	}

	defer resp.Body.Close()

	fields := map[string]interface{}{}
	err = json.Unmarshal(body, &fields)
	if err != nil {
		return types.Item{}, err
	}

	result := fields["result"].(map[string]interface{})
	if err != nil {
		r.AddField("contract", "??? "+err.Error())
	} else {
		r.AddField("contract", result["contract"])
	}

	if err != nil {
		r.AddField("network", "??? "+err.Error())
	} else {
		r.AddField("network", result["network"])
	}

	return types.Item{
		Record: r,
	}, nil
}

func (c *ZkSync) SendTransaction(ctx context.Context, from types.Account, to *common.Address, options ...interface{}) (common.Hash, error) {
	panic("implement me")
}

func (c *ZkSync) TokenInfo(ctx context.Context, token common.Address) (TokenInfo, error) {
	panic("implement me")
}

func (c *ZkSync) SendQuery(ctx context.Context, from common.Address, to common.Address, options ...interface{}) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
