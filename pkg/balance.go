package cethacea

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {

	balanceCmd := cobra.Command{
		Use:     "balance",
		Aliases: []string{"b"},
		Short:   "show balances of the current (or other) account",
	}
	all := balanceCmd.Flags().Bool("all", false, "List balance of all accounts")
	balanceCmd.RunE = func(cmd *cobra.Command, args []string) error {
		ceth, err := NewCethContext(&Settings)
		if err != nil {
			return err
		}
		dest := ""
		if len(args) > 0 {
			dest = args[0]
		}
		return balance(ceth, dest, *all)
	}
	RootCmd.AddCommand(&balanceCmd)
}

func balance(ceth *Ceth, address string, all bool) error {
	ctx := context.Background()

	c, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

	if !all {
		account, err := ceth.AccountRepo.GetCurrentAccount()
		if err != nil {
			return err
		}
		target := account.Address()
		if address != "" {
			target, err = ceth.ResolveAddress(address)
			if err != nil {
				return err
			}
		}
		balance, err := c.Balance(ctx, target)
		if err != nil {
			return errors.Wrap(err, "Couldn't get balance for")
		}
		fmt.Println(PrintEthFromDecimal(balance))
	} else {
		accounts, err := ceth.AccountRepo.ListAccounts()
		if err != nil {
			return err
		}
		for _, a := range accounts {

			balance, err := c.Balance(ctx, a.Address())
			if err != nil {
				return errors.Wrap(err, "Couldn't get balance for "+a.Address().String())
			}
			fmt.Printf("%s %s\n", a.Name, PrintEthFromDecimal(balance))
		}
	}

	return nil

}
