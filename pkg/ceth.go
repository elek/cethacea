package cethacea

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/elek/cethacea/pkg/config"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"strings"
)

type CethSettings struct {
	Contract string
	Chain    string
	Account  string
	Abi      string
	Format   string
	All      bool
	Debug    bool
}

type Ceth struct {
	AccountRepo  *config.AccountRepo
	ContractRepo *config.ContractRepo
	ChainManager *config.ChainRepo
	Settings     *CethSettings
}

func (c *Ceth) GetClient() (*chain.Eth, error) {
	cfg, err := c.ChainManager.GetCurrentChain()
	if err != nil {
		return nil, err
	}
	if cfg.Protocol != "" && cfg.Protocol != "eth" {
		return nil, fmt.Errorf("this opreation is not supported with protocol %s", cfg.Protocol)
	}
	return chain.NewEthFromURL(cfg.RPCURL)
}

func (c *Ceth) GetChainClient() (chain.ChainClient, error) {
	cfg, err := c.ChainManager.GetCurrentChain()
	if err != nil {
		return nil, err
	}
	if cfg.Protocol == "" {
		cfg.Protocol = "eth"
	}

	switch cfg.Protocol {
	case "eth":
		return chain.NewEthFromURL(cfg.RPCURL)
	case "zksync":
		return chain.NewZkSyncFromURL(cfg.RPCURL)
	default:
		return nil, fmt.Errorf("unsupported protocol %s", cfg.Protocol)
	}
}

func (c *Ceth) GetRpcClient(ctx context.Context) (*rpc.Client, error) {
	chain, err := c.ChainManager.GetCurrentChain()
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc.DialContext(ctx, chain.RPCURL)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}

func (c *Ceth) AccountContractClient() (types.Account, types.Contract, *chain.Eth, error) {
	account, err := c.AccountRepo.GetCurrentAccount()
	if err != nil {
		return types.Account{}, types.Contract{}, nil, err
	}
	contract, err := c.GetCurrentContract()
	if err != nil {
		return types.Account{}, types.Contract{}, nil, err
	}
	client, err := c.GetClient()
	if err != nil {
		return types.Account{}, types.Contract{}, nil, err
	}
	return account, contract, client, nil
}

func (c *Ceth) AccountClient() (types.Account, *chain.Eth, error) {
	account, err := c.AccountRepo.GetCurrentAccount()
	if err != nil {
		return types.Account{}, nil, err
	}

	client, err := c.GetClient()
	if err != nil {
		return types.Account{}, nil, err
	}
	return account, client, nil
}

func NewCethContext(settings *CethSettings) (*Ceth, error) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if settings.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	cm, err := config.NewChainRepo(viper.GetString("chain"))
	if err != nil {
		return nil, err
	}

	am, err := config.NewAccountRepo(viper.GetString("account"))
	if err != nil {
		return nil, err
	}

	crt, err := config.NewContractRepo(viper.GetString("contract"), viper.GetString("abi"), settings.All)
	if err != nil {
		return nil, err
	}

	return &Ceth{
		Settings:     settings,
		ChainManager: cm,
		ContractRepo: crt,
		AccountRepo:  am,
	}, nil
}

func (c *Ceth) ResolveAddress(address string) (common.Address, error) {
	if address == "" {
		account, err := c.AccountRepo.GetCurrentAccount()
		if err != nil {
			return common.Address{}, err
		}
		return account.Address(), nil
	}
	acc, err := c.AccountRepo.GetAccount(address)
	if err == nil {
		return acc.Address(), nil
	}

	contract, err := c.ContractRepo.GetContract(address)
	if err == nil {
		return contract.GetAddress(), nil
	}
	address = strings.TrimPrefix(address, "0x")
	decoded, err := hex.DecodeString(address)
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(decoded), nil
}

func (c *Ceth) GetCurrentContract() (types.Contract, error) {
	return c.ContractRepo.GetCurrentContract()
}

func (c *Ceth) getCurrentChainID() (int64, error) {
	chainCfg, err := c.ChainManager.GetCurrentChain()
	if err != nil {
		return 0, err
	}

	if chainCfg.ChainID != 0 {
		return chainCfg.ChainID, nil
	}

	client, err := c.GetChainClient()
	if err != nil {
		return 0, err
	}

	return client.GetChainID(context.Background())
}

func (c *Ceth) SetDefaultAccount(s string) error {
	c.Settings.Contract = s
	return config.ChangeKey(".ceth.yaml", "account", s)
}

func (c *Ceth) SetDefaultContract(name string) error {
	c.Settings.Contract = name
	return config.ChangeKey(".ceth.yaml", "contract", name)
}

func (c *Ceth) SetDefaultChain(s string) error {
	c.Settings.Chain = s
	return config.ChangeKey(".ceth.yaml", "chain", s)
}
