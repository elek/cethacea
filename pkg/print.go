package cethacea

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/elek/cethacea/pkg/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"strconv"
	"strings"
)

func PrintTransaction(tx *ethtypes.Transaction, receipt *ethtypes.Receipt) {
	fmt.Printf("Hash:         %s\n", tx.Hash())
	fmt.Printf("Successfull:  %d\n", receipt.Status)
	fmt.Printf("Block:        %s\n", receipt.BlockHash.Hex())

	if tx.To() != nil {
		fmt.Printf("To:           %s\n", tx.To().Hex())
	} else {
		fmt.Printf("To:           0 (CONTRACT CREATION)\n")
		fmt.Printf("New contract address: %s\n", receipt.ContractAddress)
	}
	fmt.Printf("Value:        %d\n", tx.Value())
	fmt.Printf("Gas:          %d\n", tx.Gas())
	fmt.Printf("Gas used:     %d\n", receipt.GasUsed)
	fmt.Printf("Cumulative:   %d\n", receipt.CumulativeGasUsed)
	fmt.Printf("Gas fee cap:  %d\n", tx.GasFeeCap())
	fmt.Printf("Type:         %d\n", tx.Type())
	fmt.Printf("Data:         %s\n", hex.EncodeToString(tx.Data()))
}

func PrintGWei(wei *big.Int) string {
	if wei == nil {
		return ""
	}
	return decimal.NewFromBigInt(wei, -9).String()
}

func PrintRawLog(l ethtypes.Log) {
	fmt.Println(l.Index)
	fmt.Println(l.Removed)
	for _, topic := range l.Topics {
		fmt.Println("  " + topic.Hex())
	}
	fmt.Println(hex.EncodeToString(l.Data))
	fmt.Println()
}

func PrintItem(item types.Item, format string) error {
	maxKeyLength := 0
	switch format {
	case "json":
		res := make(map[string]interface{})
		for _, record := range item.Record.Fields {
			res[record.Name] = record.Value
		}
		bytes, err := json.MarshalIndent(res, "", "   ")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	case "csv":
		var line []string
		for _, record := range item.Record.Fields {
			line = append(line, record.String())
		}
		fmt.Println(strings.Join(line, ","))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		for _, record := range item.Record.Fields {
			table.Append([]string{record.Name, wrap(record.String(), 80)})
		}
		table.Render()
		fmt.Println()
	default:

		for _, record := range item.Record.Fields {
			if len(record.Name) > maxKeyLength {
				maxKeyLength = len(record.Name)
			}
		}
		for _, record := range item.Record.Fields {
			if len(record.Name) > maxKeyLength {
				maxKeyLength = len(record.Name)
			}
			fmt.Printf("%-"+strconv.Itoa(maxKeyLength+1)+"s %s\n", record.Name+":", record.String())
		}
		fmt.Println()
	}
	return nil
}

func wrap(s string, i int) string {
	if len(s) < i {
		return s
	}
	return s[0:i] + "\n" + wrap(s[i:], i)
}
