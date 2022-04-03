package cethacea

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs/v2"
	"math/big"
	"strings"
	"time"
)

func init() {

	blockCmd := cobra.Command{
		Use:   "block",
		Short: "block related operations",
	}
	{
		blockShow := cobra.Command{
			Use:   "show <hashOrNumber>",
			Short: "Show last mined blocks",
		}
		blockShow.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			hashOrNo := ""
			if len(args) > 0 {
				hashOrNo = args[0]
			}
			return showBlock(ceth, hashOrNo)
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
	{

		blockAt := cobra.Command{
			Use:   "at",
			Short: "Find block at a specific time",
		}
		blockAt.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return findBlockAt(ceth, args[0])
		}
		blockCmd.AddCommand(&blockAt)

	}
	RootCmd.AddCommand(&blockCmd)
}

func findBlockAt(ceth *Ceth, s string) error {
	ctx := context.Background()

	point := time.Now().Add(-356 * 24 * time.Hour)

	client, err := ceth.GetChainClient()
	ethClient := client.(*chain.Eth)
	if err != nil {
		return err
	}

	first := big.NewInt(0)
	lastBlock, err := ethClient.Client.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}

	last := lastBlock.Number()

	for {
		diff := new(big.Int).Sub(last, first).Int64()
		if diff == int64(1) {
			break
		}
		log.Debug().Int64("diff", diff).Msg("binary search for block")
		middle := new(big.Int).Div(new(big.Int).Add(last, first), big.NewInt(2))
		middleBlock, err := ethClient.Client.BlockByNumber(ctx, middle)
		if err != nil {
			return err
		}
		if middleBlock.ReceivedAt.Before(point) {
			first = middleBlock.Number()
		} else {
			last = middleBlock.Number()
		}
	}
	selected, err := ethClient.Client.BlockByNumber(ctx, last)
	if err != nil {
		return err
	}
	fmt.Println(selected.Number())
	return nil
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
	for b := lastBlock; b+limit > lastBlock && b > 0; b-- {
		block, err := client.Client.BlockByNumber(ctx, big.NewInt(int64(b)))
		if err != nil {
			return errs.Wrap(err)
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

func showBlock(ceth *Ceth, hashOrNumber string) (err error) {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	var block *types.Block
	ctx := context.Background()
	if hashOrNumber == "" {
		block, err = client.Client.BlockByNumber(ctx, nil)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(hashOrNumber, "0x") {
		block, err = client.Client.BlockByHash(ctx, common.HexToHash(hashOrNumber))
		if err != nil {
			return err
		}
	} else {
		bn, ok := big.NewInt(0).SetString(hashOrNumber, 10)
		if !ok {
			return errs.Errorf("Couldn't convert to big number. To use hex please prefix it with 0x")
		}
		block, err = client.Client.BlockByNumber(ctx, bn)
		if err != nil {
			return err
		}
	}

	created := time.Unix(int64(block.Header().Time), 0)
	fmt.Printf("%-15s %s\n", "Block:", block.Hash())
	fmt.Printf("%-15s %d\n", "Block#:", block.Number())
	fmt.Printf("%-15s %s\n", "Time:", created.Format(time.RFC3339))
	fmt.Printf("%-15s %d\n", "Gas used:", block.GasUsed())
	fmt.Printf("%-15s %d\n", "Base fee:", block.BaseFee())
	fmt.Printf("%-15s %d\n", "Tx#:", len(block.Transactions()))
	fmt.Println()
	for r, tx := range block.Transactions() {
		fmt.Printf("#%3d %s %10d %19s %3d\n", r, tx.Hash(), tx.Gas(), PrintGWei(tx.GasTipCap()), tx.Type())
	}
	return nil
}
