package cethacea

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"math/big"
)

func init() {

	utilCmd := cobra.Command{
		Use:   "util",
		Short: "various converters and utilities",
	}
	RootCmd.AddCommand(&utilCmd)
	{
		hexCmd := cobra.Command{
			Use:   "hex <number>",
			Short: "Convert huge integers to hex",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				value, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("couldn't convert %s to number", args[0])
				}
				fmt.Println(value.Text(16))
				return nil
			},
		}
		utilCmd.AddCommand(&hexCmd)
	}
	{
		decCmd := cobra.Command{
			Use:   "dec <number>",
			Short: "Convert huge integers to hex",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				value, ok := new(big.Int).SetString(args[0], 16)
				if !ok {
					return fmt.Errorf("couldn't convert %s to number", args[0])
				}
				fmt.Println(value.Text(10))
				return nil
			},
		}
		utilCmd.AddCommand(&decCmd)
	}
	{
		cmd := cobra.Command{
			Use:   "keccak <number>",
			Short: "Calculate hash of keccak",
			Args:  cobra.ExactArgs(1),
		}
		input := cmd.Flags().String("input", "string", "The interpretation of input string (string,decimal,hex)")
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			var v []byte
			switch *input {
			case "hex":
				var err error
				v, err = hex.DecodeString(args[0])
				if err != nil {
					return err
				}
			case "string":
				v = []byte(args[0])
			case "decimal":
				value, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("couldn't convert %s to number", args[0])
				}
				v = make([]byte, 32)
				for i := 31; i >= 0; i-- {
					if 31-i <= len(value.Bytes())-1 {
						v[i] = value.Bytes()[len(value.Bytes())-1-31+i]
					}
				}
			}

			hash := crypto.Keccak256(v)
			fmt.Println(hex.EncodeToString(hash))
			return nil
		}
		utilCmd.AddCommand(&cmd)
	}
	{
		estimateCmd := cobra.Command{
			Use:   "estimate",
			Short: "Give gas fee / tip estimation",
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}

				client, err := ceth.GetClient()
				if err != nil {
					return err
				}
				ctx := context.Background()
				tipCap, err := client.Client.SuggestGasTipCap(ctx)
				if err != nil {
					return err
				}
				gasPrice, err := client.Client.SuggestGasPrice(ctx)
				if err != nil {
					return err
				}

				head, err := client.Client.BlockByNumber(ctx, nil)
				if err != nil {
					return err
				}
				fmt.Printf("%-25s: %s\n", "Suggested tip", PrintGWei(tipCap))
				fmt.Printf("%-25s: %s\n", "Suggested gas price", PrintGWei(gasPrice))
				fmt.Printf("%-25s: %s\n", "Base fee", PrintGWei(head.BaseFee()))
				return nil
			},
		}
		utilCmd.AddCommand(&estimateCmd)
	}

}
