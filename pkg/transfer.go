package cethacea

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/spf13/cobra"
	"math/big"
)

func init() {

	transferCmd := cobra.Command{
		Use:     "transfer <amount> <destination>",
		Aliases: []string{"t"},
		Short:   "transfer native token or token to other address",
		Args:    cobra.ExactArgs(2),
	}
	transferCmd.RunE = func(cmd *cobra.Command, args []string) error {
		ceth, err := NewCethContext(&Settings)
		if err != nil {
			return err
		}
		return nativeTransfer(ceth, args[0], args[1])
	}
	RootCmd.AddCommand(&transferCmd)
}

func nativeTransfer(ceth *Ceth, amount string, to string) error {
	account, client, err := ceth.AccountClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	value, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("invalid value")
	}
	target, err := ceth.ResolveAddress(to)
	if err != nil {
		return err
	}
	tx, err := client.SendTransaction(ctx, account, &target, chain.WithValue{Value: value})
	if err != nil {
		return err
	}
	fmt.Println(tx)
	return err
}

func tokenTransfer(ceth *Ceth, amount string, to string) error {
	ctx := context.Background()
	account, contract, client, err := ceth.AccountContractClient()
	if err != nil {
		return err
	}
	argumentTypes := abi.Arguments{
		abi.Argument{
			Name: "address",
			Type: abi.Type{
				Size: 20,
				T:    abi.AddressTy,
			},
		},
		abi.Argument{
			Name: "uint256",
			Type: abi.Type{
				Size: 256,
				T:    abi.IntTy,
			},
		},
	}

	target, err := ceth.ResolveAddress(to)
	if err != nil {
		return err
	}
	value, _ := new(big.Int).SetString(amount, 10)
	data, err := client.FunctionCallData("transfer(address,uint256)", argumentTypes, target, value)
	if err != nil {
		return err
	}

	ca := contract.GetAddress()
	tx, err := client.SendTransaction(ctx, account,
		&ca,
		chain.WithData{Data: data})
	if err != nil {
		return err
	}

	fmt.Println(tx.Hex())
	return nil
}
