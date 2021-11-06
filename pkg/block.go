package cethacea

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"math/big"
	"time"
)

func init() {

	blockCmd := cobra.Command{
		Use:   "block",
		Short: "block related operations",
	}
	{
		blockShow := cobra.Command{
			Use:   "show",
			Short: "Show last mined blocks",
		}
		blockShow.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return showBlock(ceth)
		}
		blockCmd.AddCommand(&blockShow)

	}
	{
		var limit uint64
		blockListCmd := cobra.Command{
			Use:   "list",
			Short: "List last mined blocks",
		}
		blockListCmd.Flags().Uint64Var(&limit, "limit", 10, "Number of blocks to print out")
		blockListCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return listBlocks(ceth, limit)
		}
		blockCmd.AddCommand(&blockListCmd)

	}
	{

		blockWatchCmd := cobra.Command{
			Use:   "watch",
			Short: "Watch for new head blocks",
		}
		blockWatchCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return watchBlocks(ceth)
		}
		blockCmd.AddCommand(&blockWatchCmd)

	}
	RootCmd.AddCommand(&blockCmd)
}

func watchBlocks(ceth *Ceth) error {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	ch := make(chan *types.Header)
	subscription, err := client.Client.SubscribeNewHead(ctx, ch)
	if err != nil {
		return err
	}

	select {
	case head := <-ch:
		fmt.Println(head.Number)
	case <-subscription.Err():
		break
	}
	return nil
}

func listBlocks(ceth *Ceth, limit uint64) error {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	lastBlock, err := client.Client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	for b := lastBlock; b > lastBlock-limit; b-- {
		block, err := client.Client.BlockByNumber(ctx, big.NewInt(int64(b)))
		if err != nil {
			return err
		}
		gasString := fmt.Sprintf("%10d", block.GasUsed())
		if block.GasUsed() > 15_000_000 {
			gasString = color.RedString("%10d", block.GasUsed())
		}

		var maxTip, minTip, sumTip *big.Int
		sumTip = big.NewInt(0)
		for _, tx := range block.Transactions() {
			tip := tx.GasTipCap()
			if maxTip == nil || tip.Cmp(maxTip) > 0 {
				maxTip = tip
			}
			if minTip == nil || tip.Cmp(minTip) < 0 {
				minTip = tip
			}
			sumTip = sumTip.Add(sumTip, tip)

		}
		blockTime := time.Unix(int64(block.Time()), 0)
		average := big.NewInt(0)
		if len(block.Transactions()) > 0 {
			average = new(big.Int).Div(sumTip, big.NewInt(int64(len(block.Transactions()))))
		}
		fmt.Printf("%d %10s %s %s %18s %3d %s/%s/%s\n",
			block.Number(),
			blockTime.Format("2006-01-02T15:04:05"),
			block.Hash().String(),
			gasString,
			PrintGWei(block.BaseFee()),
			len(block.Transactions()),
			PrintGWei(minTip),
			PrintGWei(average),
			PrintGWei(maxTip),
		)
	}
	return nil

}

func showBlock(ceth *Ceth) error {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	block, err := client.Client.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("%-15s %s\n", "Block:", block.Hash())
	fmt.Printf("%-15s %d\n", "Block#:", block.Number())
	fmt.Printf("%-15s %s\n", "Received at:", block.ReceivedAt.Format(time.RFC3339))
	fmt.Printf("%-15s %d\n", "Gas used:", block.GasUsed())
	fmt.Printf("%-15s %d\n", "Base fee:", block.BaseFee())
	fmt.Printf("%-15s %d\n", "Tx#:", len(block.Transactions()))
	fmt.Println()
	for r, tx := range block.Transactions() {
		fmt.Printf("#%3d %s %10d %19s %3d\n", r, tx.Hash(), tx.Gas(), PrintGWei(tx.GasTipCap()), tx.Type())
	}
	return nil
}
