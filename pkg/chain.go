package cethacea

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/config"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"strings"
)

func init() {

	chainCmd := cobra.Command{
		Use:     "chain",
		Aliases: []string{"ch"},
		Short:   "Configure current RPC endpoint",
	}

	addCmd := cobra.Command{
		Use:  "add <name> <url>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return addChain(ceth, args[0], args[1])
		},
	}
	listCmd := cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return listChain(ceth)
		},
	}
	{
		idCmd := cobra.Command{
			Use: "id",
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}
				return id(ceth)
			},
		}
		chainCmd.AddCommand(&idCmd)
	}
	{
		cmd := cobra.Command{
			Use: "info",
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}
				return info(ceth)
			},
		}
		chainCmd.AddCommand(&cmd)
	}
	switchCmd := cobra.Command{
		Use:     "switch <name>",
		Aliases: []string{"s", "sw"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			selected := ""
			if len(args) > 0 {
				selected = args[0]
			}
			return switchChain(ceth, selected)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			chm, err := config.NewChainRepo("")
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			var a []string
			chains, err := chm.ListChains()
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			for _, c := range chains {
				if strings.HasPrefix(c.Name, toComplete) {
				}
				a = append(a, c.Name)
			}
			return a, cobra.ShellCompDirectiveNoFileComp
		},
	}
	chainCmd.AddCommand(&addCmd)
	chainCmd.AddCommand(&listCmd)
	chainCmd.AddCommand(&switchCmd)
	RootCmd.AddCommand(&chainCmd)
}

func info(ceth *Ceth) error {
	client, err := ceth.GetChainClient()
	if err != nil {
		return err
	}
	info, err := client.GetChainInfo(context.Background())
	if err != nil {
		return err
	}
	return PrintItem(info, ceth.Settings.Format)
}

func id(ceth *Ceth) error {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	id, err := client.Client.NetworkID(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("NetworkID: " + id.String())
	chainId, err := client.Client.ChainID(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("ChainID: " + chainId.String())
	return nil
}

func switchChain(ceth *Ceth, s string) error {
	if s == "" {
		chains, err := ceth.ChainManager.ListChains()
		if err != nil {
			return err
		}

		idx, _ := fuzzyfinder.Find(chains, func(i int) string {
			return fmt.Sprintf("[%s] %s", chains[i].Name, chains[i].RPCURL)
		})
		err = ceth.SetDefaultChain(chains[idx].Name)
		if err != nil {
			return err
		}
	} else {
		err := ceth.SetDefaultChain(s)
		if err != nil {
			return err
		}
	}
	return listChain(ceth)
}

func listChain(ceth *Ceth) error {
	chains, err := ceth.ChainManager.ListChains()
	if err != nil {
		return err
	}
	def, err := ceth.ChainManager.GetCurrentChain()
	if err != nil {
		return err
	}

	for _, c := range chains {
		marker := " "
		if c.Name ==
			def.Name {
			marker = "*"
		}
		fmt.Println(marker + c.Name + " " + c.RPCURL)
	}
	return nil
}

func addChain(ceth *Ceth, name string, url string) error {
	return ceth.ChainManager.AddChain(types.ChainConfig{
		Name:   name,
		RPCURL: url,
	})
}
