package config

import (
	"github.com/elek/cethacea/pkg/types"
	"github.com/pkg/errors"
	"os"
	"os/user"
	"path"
	"strings"
)

type ChainRepo struct {
	Chains     []types.ChainConfig
	ConfigFile string
	Selected   string
}

func NewChainRepo(selected string) (*ChainRepo, error) {

	config, err := globalChainConfig()
	if err != nil {
		return nil, err
	}
	var chains []types.ChainConfig
	err = LoadYamlConfig(config, "chains", &chains)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(selected, "http") || strings.HasPrefix(selected, "ws") {
		chains = append(chains, types.ChainConfig{
			Name:   "<chain>",
			RPCURL: selected,
		})

		return &ChainRepo{
			Chains:     chains,
			Selected:   "<chain>",
			ConfigFile: config,
		}, nil
	}

	chainEnv := os.Getenv("CETH_CHAIN")
	if chainEnv != "" {
		if strings.HasPrefix(chainEnv, "http") || strings.HasPrefix(chainEnv, "ws") {
			chains = append(chains, types.ChainConfig{
				Name:   "CETH_CHAIN",
				RPCURL: chainEnv,
			})
			selected = "CETH_CHAIN"
		}
		for _, c := range chains {
			if c.Name == chainEnv {
				selected = chainEnv
			}
		}
	}

	if selected == "" {
		selected = "default"
	}
	return &ChainRepo{
		Chains:     chains,
		Selected:   selected,
		ConfigFile: config,
	}, nil
}

func (c *ChainRepo) ListChains() ([]types.ChainConfig, error) {
	return c.Chains, nil
}

func (c *ChainRepo) GetCurrentChain() (types.ChainConfig, error) {

	for _, chain := range c.Chains {
		if c.Selected == chain.Name {
			return chain, nil
		}
	}
	moreHelp := ""
	if c.Selected == "default" {
		moreHelp = " Please configure your ethereum RPC endpoint with `cethacea chain add default https://...."
	}
	return types.ChainConfig{}, errors.New("Selected chain " + c.Selected + " is not found." + moreHelp)

}
func (c *ChainRepo) AddChain(config types.ChainConfig) error {
	var chains []types.ChainConfig
	err := LoadYamlConfig(c.ConfigFile, "chains", &chains)
	if err != nil {
		return err
	}
	chains = append(chains, config)
	return SaveYamlConfig(c.ConfigFile, &chains)
}

func globalChainConfig() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		configHome = path.Join(usr.HomeDir, ".config")
	}
	chainConfig := path.Join(configHome, "cethacea", "chains.yaml")
	_ = os.MkdirAll(path.Dir(chainConfig), 0700)
	return chainConfig, nil
}
