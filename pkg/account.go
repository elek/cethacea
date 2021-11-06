package cethacea

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/config"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var DefaultAccountFileName = ".accounts.yaml"

func init() {

	accountCmd := cobra.Command{
		Use:     "account",
		Short:   "account related operations",
		Aliases: []string{"a", "key", "keys"},
	}

	generateCmd := cobra.Command{
		Use:     "generate <name>",
		Aliases: []string{"g"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := config.NewAccountRepo("")
			if err != nil {
				return err
			}
			return generateAccount(f)
		},
	}
	{
		listCmd := cobra.Command{
			Use:     "list <name>",
			Aliases: []string{"l"},
			Args:    cobra.MaximumNArgs(1),
		}
		all := listCmd.Flags().Bool("all", false, "Show all information including the private key")
		listCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return listAccounts(ceth.AccountRepo, *all)
		}
		accountCmd.AddCommand(&listCmd)
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
			return switchAccount(ceth, selected)
		},
	}
	{
		infoCmd := cobra.Command{
			Use:     "info",
			Aliases: []string{"i"},
			Short:   "Show information about accounts",
		}
		infoCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}

			return accountInfo(ceth)
		}
		accountCmd.AddCommand(&infoCmd)
	}
	accountCmd.AddCommand(&generateCmd)
	accountCmd.AddCommand(&switchCmd)
	RootCmd.AddCommand(&accountCmd)
}

func accountInfo(ceth *Ceth) error {
	ctx := context.Background()

	c, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

	account, err := ceth.AccountRepo.GetCurrentAccount()
	if err != nil {
		return err
	}
	target := account.Address()

	info, err := c.GetAccountInfo(ctx, target)
	if err != nil {
		return errors.Wrap(err, "Couldn't get balance for")
	}
	return PrintItem(info, ceth.Settings.Format)
}

func switchAccount(ceth *Ceth, s string) error {
	f := ceth.AccountRepo
	if s == "" {
		accounts, err := f.ListAccounts()
		if err != nil {
			return nil
		}

		idx, _ := fuzzyfinder.Find(accounts, func(i int) string {
			return fmt.Sprintf("[%s] %s", accounts[i].Name, accounts[i].Address())
		})
		err = ceth.SetDefaultAccount(accounts[idx].Name)
		if err != nil {
			return nil
		}
	} else {
		err := ceth.SetDefaultAccount(s)
		if err != nil {
			return nil
		}
	}
	return listAccounts(f, false)
}

func generateAccount(f *config.AccountRepo) error {
	name := f.GetNextName()
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return errors.Wrap(err, "Couldn't generate random private key")
	}

	a := types.Account{
		Name:    name,
		Private: hex.EncodeToString(crypto.FromECDSA(key)),
	}
	err = f.AddAccount(a)
	return err
}

func listAccounts(f *config.AccountRepo, all bool) error {
	accounts, err := f.ListAccounts()
	if err != nil {
		return err
	}
	curr, _ := f.GetCurrentAccount()
	for _, a := range accounts {
		marker := " "
		if a.Name == curr.Name {
			marker = "*"
		}
		if all {
			fmt.Println(marker + a.Name + " " + a.Address().Hex() + " " + a.Private)
		} else {
			fmt.Println(marker + a.Name + " " + a.Address().Hex())
		}
	}
	return nil
}
