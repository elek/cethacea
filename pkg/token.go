package cethacea

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	"math/big"
)

func init() {

	tokenCmd := cobra.Command{
		Use:   "token",
		Short: "ERC20 token related operations",
	}
	RootCmd.AddCommand(&tokenCmd)
	{
		cmd := cobra.Command{
			Use:   "balance",
			Short: "Show balance in the selected token",
			Args:  cobra.MaximumNArgs(1),
		}
		all := cmd.Flags().Bool("all", false, "List balance of all accounts")
		raw := cmd.Flags().Bool("raw", false, "Print only raw value")
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			dest := ""
			if len(args) > 0 {
				dest = args[0]
			}
			return tokenBalance(ceth, dest, *raw, *all)
		}
		tokenCmd.AddCommand(&cmd)
	}
	{
		cmd := cobra.Command{
			Use:   "info",
			Short: "Show basic information from the token",
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return tokenInfo(ceth, Settings.Format)
		}
		tokenCmd.AddCommand(&cmd)
	}

}

func tokenInfo(ceth *Ceth, format string) error {
	c, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	info, err := c.TokenInfo(context.Background(), contract.GetAddress())
	if err != nil {
		return err
	}
	return PrintItem(info, format)
}

func tokenBalance(ceth *Ceth, address string, raw bool, all bool) error {
	ctx := context.Background()
	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	c, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

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

	decimals := uint8(0)
	symbol := ""

	if !raw {
		info, err := c.TokenInfo(ctx, contract.GetAddress())
		if err != nil {
			return err
		}
		decimals = info.GetUint8("decimals")
		symbol = info.GetString("symbol")
	}

	amount, err := c.TokenBalance(ctx, contract.GetAddress(), target)
	if err != nil {
		return err
	}

	fmt.Println(PrintAmount(amount, decimals, symbol))
	return nil
}

func PrintAmount(amount *big.Int, dec uint8, symbol string) string {
	out := amount.String()
	if dec != 0 && symbol != "" {
		out = decimal.NewFromBigInt(amount, -1*int32(dec)).String()
		out += " " + symbol
	}
	return out

}
