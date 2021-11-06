package cethacea

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/elek/cethacea/pkg/encoding"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"math/big"
	"time"
)

func init() {

	txCommand := cobra.Command{
		Use:   "tx",
		Short: "transaction related operations",
	}
	{
		txGetCommand := cobra.Command{
			Use:   "get <tx>",
			Short: "Get information about a specific transaction",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}
				return showTx(ceth, args[0], Settings.Format)
			},
		}
		txCommand.AddCommand(&txGetCommand)
	}
	{
		txDebugCommand := cobra.Command{
			Use:   "debug <tx>",
			Short: "Show debug information of a given transaction (if available)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}

				return debugTx(ceth, args[0])
			},
		}
		txCommand.AddCommand(&txDebugCommand)
	}
	{
		txCancelCmd := cobra.Command{
			Use:   "cancel <tx>",
			Short: "Cancel pending transaction",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ceth, err := NewCethContext(&Settings)
				if err != nil {
					return err
				}

				return cancelTx(ceth, args[0])
			},
		}
		txCommand.AddCommand(&txCancelCmd)
	}
	{
		txSubmitCmd := cobra.Command{
			Use:   "submit <tx>",
			Short: "Submit raw transaction",
		}

		value := txSubmitCmd.Flags().String("value", "", "Value of the transaction")
		data := txSubmitCmd.Flags().String("data", "", "Hex data of the transaction")
		to := txSubmitCmd.Flags().String("to", "", "Target address of the transaction")
		txSubmitCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}

			return submit(ceth, *value, *to, *data)
		}
		txCommand.AddCommand(&txSubmitCmd)
	}
	RootCmd.AddCommand(&txCommand)
}

func debugTx(ceth *Ceth, s string) error {
	ctx := context.Background()

	client, err := ceth.GetRpcClient(ctx)
	if err != nil {
		return err
	}
	k := map[string]interface{}{
		"EnableMemory": true,
	}
	res := map[string]interface{}{}
	err = client.CallContext(ctx, &res, "debug_traceTransaction", s, k)
	if err != nil {
		return err
	}
	fmt.Printf("Failed:    %v\n", res["failed"].(bool))
	fmt.Printf("Gas:       %v\n", res["gas"].(float64))
	fmt.Printf("Ret:       %v\n", res["returnValue"].(string))
	fmt.Println()
	fmt.Println("Stack top is on right")
	fmt.Println()
	for _, log := range res["structLogs"].([]interface{}) {
		record := log.(map[string]interface{})
		fmt.Printf("%-5d %-14s %-2d %-2d %v\n",
			int(record["pc"].(float64)),
			record["op"],
			int(record["gasCost"].(float64)),
			int(record["depth"].(float64)),
			record["stack"],
		)
		for _, m := range record["memory"].([]interface{}) {
			fmt.Printf("%91s\n", m)
		}
	}
	return err
}

func cancelTx(ceth *Ceth, s string) error {
	account, client, err := ceth.AccountClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	txHash := common.HexToHash(s)
	tx, pending, err := client.Client.TransactionByHash(ctx, txHash)
	if !pending {
		return fmt.Errorf("transaction %s is not in pending state", txHash)
	}
	if err != nil {
		return err
	}

	to := account.Address()
	res, err := client.SubmitTx(ctx, account, &to,
		chain.WithGas{Gas: tx.Gas() + 10},
		chain.WithNonce{Nonce: tx.Nonce()},
		chain.WithGasFeeCap{Value: new(big.Int).Add(tx.GasTipCap(), big.NewInt(10))},
		chain.WithGasTipCap{new(big.Int).Add(tx.GasTipCap(), big.NewInt(10))})
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func showTx(ceth *Ceth, s string, format string) error {
	ctx := context.Background()
	hash := common.HexToHash(s)
	c, err := ceth.GetChainClient()
	if err != nil {
		return err
	}
	tx, err := c.GetTransaction(ctx, hash)
	if err != nil {
		return err
	}
	return PrintItem(tx, format)
}

func submit(ceth *Ceth, value string, to string, data string) error {
	ctx := context.Background()

	client, err := ceth.GetChainClient()
	if err != nil {
		return err
	}

	account, err := ceth.AccountRepo.GetCurrentAccount()
	if err != nil {
		return err
	}

	var toAddress *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddress = &addr
	}
	var opts []interface{}
	if data != "" {
		hexData, err := encoding.HexToBytes(data)
		if err != nil {
			return err
		}
		opts = append(opts, chain.WithData{Data: hexData})
	}
	if value != "" {
		v := new(big.Int)
		v.SetString(value, 10)
		opts = append(opts, chain.WithValue{
			Value: v,
		})
	}
	tx, err := client.SendTransaction(ctx, account, toAddress, opts...)
	if err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		info, err := client.GetTransaction(ctx, tx)
		if err == nil {
			return PrintItem(info, "console")
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Submitted transaction couldn't been retrieved")

}
