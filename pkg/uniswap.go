package cethacea

import (
	"context"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
)

const pairsFile = "pairsv2.yaml"

func init() {

	uniswap := cobra.Command{
		Use:     "uniswap2",
		Aliases: []string{"uv2"},
		Short:   "cli methods for uniswap manipulations",
	}
	RootCmd.AddCommand(&uniswap)
	{
		cmd := cobra.Command{
			Use:   "download",
			Short: "Download uniswap pair information",
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return uniswapV2Download(ceth)
		}
		uniswap.AddCommand(&cmd)
	}
	{
		cmd := cobra.Command{
			Use:   "price",
			Short: "Get actual price for one pair",
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return uniswapV2Price(ceth)
		}
		uniswap.AddCommand(&cmd)
	}

}

type UniPair struct {
	Contract string
	Token0   string
	Token1   string
	Symbol0  string
	Symbol1  string
	Name     string
}

func uniswapV2Price(ceth *Ceth) error {
	ctx := context.Background()

	acc, err := ceth.AccountRepo.GetCurrentAccount()
	if err != nil {
		return err
	}

	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	ethClient, err := ceth.GetClient()
	if err != nil {
		return err
	}

	token0Address, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), contract.GetAddress(), "token0()address")
	if err != nil {
		return err
	}

	token0Symbol, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), token0Address[0].(common.Address), "symbol()string")
	if err != nil {
		return err
	}

	decimals0, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), token0Address[0].(common.Address), "decimals()int8")
	if err != nil {
		return err
	}

	token1Address, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), contract.GetAddress(), "token1()address")
	if err != nil {
		return err
	}

	token1Symbol, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), token1Address[0].(common.Address), "symbol()string")
	if err != nil {
		return err
	}

	decimals1, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), token1Address[0].(common.Address), "decimals()int8")
	if err != nil {
		return err
	}

	res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), contract.GetAddress(), "getReserves()uint112,uint112,uint32")
	if err != nil {
		return err
	}
	decimals0Int := int32(decimals0[0].(int8))
	decimals1Int := int32(decimals1[0].(int8))
	numerator := decimal.NewFromBigInt(res[0].(*big.Int), -decimals0Int)
	denominator := decimal.NewFromBigInt(res[1].(*big.Int), -decimals1Int)
	result := numerator.Div(denominator)
	fmt.Printf("%s %s = %s %s\n", "1", token1Symbol[0].(string), result.String(), token0Symbol[0].(string))
	return nil
}

func uniswapV2Download(ceth *Ceth) error {
	pairs := make(map[int]*UniPair)

	if _, err := os.Stat(pairsFile); err == nil {
		content, err := ioutil.ReadFile(pairsFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(content, &pairs)
		if err != nil {
			return err
		}
	}

	defer func() {
		err := persistPairs(pairs)
		if err != nil {
			fmt.Println(err)
		}
	}()
	ctx := context.Background()

	ethClient, err := ceth.GetClient()
	if err != nil {
		return err
	}

	acc, err := ceth.AccountRepo.GetCurrentAccount()
	if err != nil {
		return err
	}

	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), contract.GetAddress(), "allPairsLength()uint256")
	if err != nil {
		return err
	}
	pairLength := int(res[0].(*big.Int).Int64())
	for i := 0; i < pairLength; i++ {
		pair, found := pairs[i]
		if !found {
			res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), contract.GetAddress(), "allPairs(uint256)address", strconv.Itoa(i))
			if err != nil {
				return err
			}

			pair = &UniPair{
				Contract: res[0].(common.Address).String(),
			}
			pairs[i] = pair
		}

		if pair.Token0 == "" {
			res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), common.HexToAddress(pair.Contract), "token0()address")
			if err != nil {
				return err
			}

			pair.Token0 = res[0].(common.Address).String()
		}

		if pair.Token1 == "" {
			res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), common.HexToAddress(pair.Contract), "token1()address")
			if err != nil {
				return err
			}
			pair.Token1 = res[0].(common.Address).String()
		}

		if pair.Symbol0 == "" {
			res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), common.HexToAddress(pair.Token0), "symbol()string")
			if err != nil {
				//return errors.Wrap(err, "")
			} else {
				pair.Symbol0 = res[0].(string)
			}
		}

		if pair.Symbol1 == "" {
			res, err := ethClient.Query(ctx, types.WithoutAddressResolution{}, acc.Address(), common.HexToAddress(pair.Token1), "symbol()string")
			if err != nil {
				//return errors.Wrap(err, "")
			} else {
				pair.Symbol1 = res[0].(string)
			}
		}

		if pair.Name == "" {
			pair.Name = pair.Symbol0 + pair.Symbol1
		}
		if i%10 == 0 {
			err = persistPairs(pairs)
			if err != nil {
				return err
			}
		}
		fmt.Println(i)
	}
	fmt.Println(pairLength)
	return nil
}

func persistPairs(pairs map[int]*UniPair) error {
	content, err := yaml.Marshal(pairs)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("pairsv2.yaml", content, 0644)
	if err != nil {
		return err
	}
	return nil
}
