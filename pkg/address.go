package cethacea

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

func init() {

	cmd := cobra.Command{
		Use:     "address <private>",
		Aliases: []string{"t"},
		Short:   "Generate address from private key",
		Args:    cobra.ExactArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		pk, err := crypto.HexToECDSA(args[0])
		if err != nil {
			return err
		}

		fmt.Println(crypto.PubkeyToAddress(pk.PublicKey).Hex())
		return nil
	}
	RootCmd.AddCommand(&cmd)
}
