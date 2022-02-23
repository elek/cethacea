package cethacea

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/elek/cethacea/pkg/chain"
	"github.com/elek/cethacea/pkg/config"
	"github.com/elek/cethacea/pkg/encoding"
	"github.com/elek/cethacea/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/big"
	"strings"
	"time"
)

func init() {

	contractCmd := cobra.Command{
		Use:     "contract",
		Short:   "contract related operations",
		Aliases: []string{"c"},
	}
	{
		deployCmd := cobra.Command{
			Use:     "deploy <file>",
			Aliases: []string{"d"},
			Short:   "Deploy new file to a new address",
		}
		raw := deployCmd.Flags().Bool("raw", false, "Use parameter as raw value")
		file := deployCmd.Flags().StringP("file", "f", "", "File where the data value is read from")
		value := deployCmd.Flags().String("value", "", "Value to send with the transaction")
		contractAlias := deployCmd.Flags().String("name", "", "Local alias to the contract to be persisted with the address.")
		deployCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			var data []byte
			if len(args) > 1 {
				_, data, err = parseInputData(ceth, raw, file, args[1:])
				if err != nil {
					return err
				}
			}
			return deploy(ceth, contractAlias, value, args[0], data)
		}
		contractCmd.AddCommand(&deployCmd)

	}
	{
		dataCmd := cobra.Command{
			Use:   "data <index>",
			Short: "Get data from the contract's memory slot",
		}
		dataCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}

			return dataSlot(ceth, args[0])
		}
		contractCmd.AddCommand(&dataCmd)

	}
	{
		codeCmd := cobra.Command{
			Use:   "code",
			Short: "Get deployed code of the contract",
		}
		codeCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}

			return code(ceth)
		}
		contractCmd.AddCommand(&codeCmd)

	}
	{
		callCmd := cobra.Command{
			Use:     "call <function> <param1> <param2> ...",
			Short:   "Call function on a deployed contract (transaction)",
			Aliases: []string{"d"},
		}
		raw := callCmd.Flags().Bool("raw", false, "Use parameter as raw value")
		file := callCmd.Flags().StringP("file", "f", "", "File where the data value is read from")
		value := callCmd.Flags().String("value", "", "Value to send with the transaction")
		callCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}

			val := big.NewInt(0)
			if *value != "" {
				val.SetString(*value, 0)
			}
			_, data, err := parseInputData(ceth, raw, file, args)
			if err != nil {
				return err
			}
			return call(ceth, val, data)
		}
		contractCmd.AddCommand(&callCmd)

	}
	{
		queryCommand := cobra.Command{
			Use:     "query <function> <param1> <param2> ...",
			Short:   "Execute a read-only call on the contract without committing transactions",
			Aliases: []string{"q"},
		}
		rawQuery := queryCommand.Flags().Bool("raw", false, "Use parameter as raw value")
		fileQuery := queryCommand.Flags().StringP("file", "f", "", "File where the data value is read from")

		queryCommand.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			fs, data, err := parseInputData(ceth, rawQuery, fileQuery, args)
			if err != nil {
				return err
			}
			return query(ceth, fs, data)
		}
		contractCmd.AddCommand(&queryCommand)
	}
	{
		logCommand := cobra.Command{
			Use:     "log",
			Aliases: []string{"logs"},
			Short:   "Print out blockchain log entries",
		}
		limit := logCommand.Flags().Uint64("limit", 0, "Limit number of blocks user to get events")
		all := logCommand.Flags().Bool("all", false, "Go back in the time until origin and download all pages one by one")
		raw := logCommand.Flags().Bool("raw", false, "Print out raw log content even if ABI is specified")
		format := logCommand.Flags().String("format", "", "Format of the output (empty or csv")
		topic0 := logCommand.Flags().String("topic0", "", "Filter for topic0")
		topic1 := logCommand.Flags().String("topic1", "", "Filter for topic1")
		topic2 := logCommand.Flags().String("topic2", "", "Filter for topic2")
		topic3 := logCommand.Flags().String("topic3", "", "Filter for topic3")

		logCommand.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return listLogs(ceth, *limit, *raw, *all, *format, *topic0, *topic1, *topic2, *topic3)
		}
		contractCmd.AddCommand(&logCommand)

	}
	listCmd := cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List managed contracts (address aliases)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return listContracts(ceth, Settings.All)
		},
	}
	switchCmd := cobra.Command{
		Use:     "switch <default>",
		Aliases: []string{"select", "s", "sw"},
		Short:   "Select default contract from the alias list",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			selected := ""
			if len(args) > 0 {
				selected = args[0]
			}
			return switchContract(ceth, selected, Settings.All)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			cm, err := config.NewContractRepo("", "", true)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			a := []string{}
			chains, err := cm.ListContracts()
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			for _, c := range chains {
				if strings.HasPrefix(c.Name, toComplete) {
					a = append(a, c.Name)
				}
			}
			return a, cobra.ShellCompDirectiveNoFileComp
		},
	}

	{

		addCmd := cobra.Command{
			Use:   "add  <name> <address>",
			Short: "Add contract address to the alias list",
		}
		abi := addCmd.Flags().String("abi", "", "Name of the abi file or <type>")
		addCmd.RunE = func(cmd *cobra.Command, args []string) error {
			ctr, err := config.NewContractRepo("", "", false)
			if err != nil {
				return err
			}
			return addContract(ctr, abi, args[0], args[1])
		}
		contractCmd.AddCommand(&addCmd)

	}
	methodCmd := cobra.Command{
		Use:     "method",
		Aliases: []string{"methods", "m"},
		Short:   "Show available methods based on the the configured ABI",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctr, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			return methods(ctr)
		},
	}
	contractCmd.AddCommand(&listCmd)
	contractCmd.AddCommand(&switchCmd)
	contractCmd.AddCommand(&methodCmd)
	RootCmd.AddCommand(&contractCmd)
}

func code(ceth *Ceth) error {
	client, err := ceth.GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}
	code, err := client.Client.CodeAt(ctx, contract.GetAddress(), nil)
	if err != nil {
		return err
	}
	fmt.Println(hex.EncodeToString(code))
	return nil
}

func dataSlot(ceth *Ceth, s string) error {
	_, contract, client, err := ceth.AccountContractClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	hash := common.HexToHash(s)
	res, err := client.Client.StorageAt(ctx, contract.GetAddress(), hash, nil)
	if err != nil {
		return err
	}
	fmt.Println(hex.EncodeToString(res))
	return nil
}

func parseInputData(ceth *Ceth, raw *bool, file *string, args []string) (fs encoding.FunctionSignature, data []byte, err error) {
	if *raw {
		data, err = hex.DecodeString(args[0])
		return fs, data, err
	} else if file != nil && *file != "" {
		content, err := ioutil.ReadFile(*file)
		if err != nil {
			return fs, data, err
		}
		hexData := string(content)
		hexData = strings.ReplaceAll(hexData, "\n", "")
		hexData = strings.ReplaceAll(hexData, " ", "")
		data, err = hex.DecodeString(strings.TrimSpace(hexData))
		if err != nil {
			return fs, data, err
		}
	} else if len(args) == 0 {
		return fs, data, nil
	} else if strings.Contains(args[0], "(") {
		fs, err := encoding.ParseFunctionSignature(args[0])
		if err != nil {
			return fs, data, err
		}
		data, err = fs.EncodeFuncCall(ceth, args[1:]...)
		return fs, data, err
	} else {
		fs.Name = args[0]
		contract, err := ceth.GetCurrentContract()
		if err != nil {
			return fs, data, err
		}

		contractAbi, err := contract.GetAbi()
		if err != nil {
			return fs, data, err
		}

		var method abi.Method
		var found bool
		if args[0] == "" {
			method = contractAbi.Constructor
		} else {
			method, found = contractAbi.Methods[args[0]]
			if !found {
				return fs, data, fmt.Errorf("no such method %s in abi %s", method, contract.Abi)
			}
		}
		fs = encoding.FunctionSignature(method)
		data, err = fs.EncodeFuncCall(ceth, args[1:]...)
		if err != nil {
			return fs, data, err
		}

		return fs, data, err

	}
	return fs, data, err
}

func methods(ctr *Ceth) error {
	contract, err := ctr.GetCurrentContract()
	if err != nil {
		return err
	}
	if contract.Abi == "" {
		return errors.New("ABI is not defined for this contract. Please define it with --abi (or in the .contracts.yaml)")
	}
	abi, err := contract.GetAbi()
	if err != nil {
		return err
	}
	for _, d := range abi.Methods {
		inputParams := []string{}
		for _, p := range d.Inputs {
			paramString := p.Name + " " + color.RedString(p.Type.String())
			inputParams = append(inputParams, strings.TrimSpace(paramString))
		}

		outputParams := []string{}
		for _, p := range d.Outputs {
			paramString := p.Name + " " + color.RedString(p.Type.String())
			outputParams = append(outputParams, strings.TrimSpace(paramString))
		}
		modifiers := []byte{' ', ' ', ' '}
		methodColor := color.GreenString
		if !d.IsConstant() {
			modifiers[1] = '*'
			methodColor = color.YellowString
		}
		fmt.Printf("%s %s(%s)%s\n", modifiers, methodColor("%s", d.RawName), strings.Join(inputParams, ","), strings.Join(outputParams, ","))
	}
	return nil

}

func addContract(ctr *config.ContractRepo, abi *string, s string, s2 string) error {
	return ctr.AddContract(types.Contract{
		Name:    s,
		Address: s2,
		Abi:     *abi,
	})
}

func switchContract(ceth *Ceth, s string, all bool) error {
	ctr := ceth.ContractRepo
	if s == "" {
		contracts, err := ctr.ListContracts()
		if err != nil {
			return err
		}
		chainID, err := ceth.getCurrentChainID()
		if err != nil {
			return err
		}
		var filtered []types.Contract
		for _, c := range contracts {
			if all || (c.ChainID == 0 || c.ChainID == chainID) {
				filtered = append(filtered, c)
			}
		}
		idx, _ := fuzzyfinder.Find(filtered, func(i int) string {
			return fmt.Sprintf("[%s] %s", filtered[i].Name, filtered[i].Address)

		})
		err = ceth.SetDefaultContract(contracts[idx].Name)
		if err != nil {
			return err
		}
	} else {
		err := ceth.SetDefaultContract(s)
		if err != nil {
			return err
		}
	}
	return listContracts(ceth, all)
}

func listContracts(ceth *Ceth, all bool) error {
	contracts, err := ceth.ContractRepo.ListContracts()
	if err != nil {
		return err
	}
	curr := ceth.Settings.Contract

	chainID, err := ceth.getCurrentChainID()
	if err != nil {
		return err
	}

	for _, a := range contracts {
		marker := " "
		if a.Name == curr {
			marker = "*"
		}
		if all || (a.ChainID == 0 || a.ChainID == chainID) {
			fmt.Println(marker + a.Name + " " + a.Address)
		}
	}
	return nil
}

func listLogs(ceth *Ceth, limit uint64, raw bool, all bool, format string, topic0 string, topic1 string, topic2 string, topic3 string) error {
	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	c, err := ceth.GetClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	head, err := c.Client.BlockNumber(ctx)
	if err != nil {
		return nil
	}

	until := head
	for {
		q := ethereum.FilterQuery{
			Addresses: []common.Address{
				contract.GetAddress(),
			},
		}
		if limit != 0 {
			q.ToBlock = big.NewInt(int64(until))
			q.FromBlock = big.NewInt(int64(until - limit))
		}

		q.Topics = make([][]common.Hash, 4)
		if topic0 != "" {
			q.Topics[0] = []common.Hash{common.HexToHash(topic0)}
		}
		if topic1 != "" {
			q.Topics[1] = []common.Hash{common.HexToHash(topic1)}
		}
		if topic2 != "" {
			q.Topics[2] = []common.Hash{common.HexToHash(topic2)}
		}
		if topic3 != "" {
			q.Topics[3] = []common.Hash{common.HexToHash(topic3)}
		}

		logs, err := c.Client.FilterLogs(ctx, q)
		if err != nil {
			return errors.Wrap(err, "Couldn't get logs for contract "+contract.Address)
		}
		contractAbi, err := contract.GetAbi()
		if !raw && err == nil {

			indexedEvents := map[common.Hash]abi.Event{}
			for _, e := range contractAbi.Events {
				indexedEvents[e.ID] = e
			}

			for _, l := range logs {
				event := indexedEvents[l.Topics[0]]
				item, err := LogAsTypedItem(event, l)
				if err != nil {
					return err
				}
				err = PrintItem(item, format)
				if err != nil {
					return err
				}
			}

		} else {
			for _, l := range logs {
				i := types.Record{
					Fields: []types.Field{},
				}
				i.AddField("index", l.TxIndex)
				i.AddField("removed", l.Removed)
				i.AddField("data", hex.EncodeToString(l.Data))

				for ix, topic := range l.Topics {
					i.AddField(fmt.Sprintf("topic%d", ix), topic.Hex())
				}
				err := PrintItem(types.Item{
					Record: i,
				}, format)
				if err != nil {
					return err
				}
			}
		}
		until = until - limit - 1
		if !all {
			break
		}
	}
	return nil
}

func LogAsTypedItem(event abi.Event, l ethtypes.Log) (types.Item, error) {
	i := types.Item{
		Record: types.Record{
			Fields: []types.Field{},
		},
	}

	i.AddField("block", l.BlockNumber)
	i.AddField("index", l.TxIndex)
	i.AddField("tx", l.TxHash)
	i.AddField("removed", l.Removed)
	i.AddField("event", event.Name)

	indexed := abi.Arguments{}
	nonIndexed := abi.Arguments{}
	var indexedNames []string
	var valueNames []string
	for _, i := range event.Inputs {
		if i.Indexed {
			i.Indexed = false
			indexed = append(indexed, i)
			indexedNames = append(indexedNames, i.Name)
		} else {
			nonIndexed = append(nonIndexed, i)
			valueNames = append(valueNames, i.Name)
		}
	}

	var topicBytes []byte
	for ix, topic := range l.Topics {
		if ix != 0 {
			topicBytes = append(topicBytes, topic.Bytes()...)
		}
	}

	topicValues, err := indexed.UnpackValues(topicBytes)
	if err != nil {
		return i, err
	}
	for ix, v := range topicValues {
		i.AddField(indexedNames[ix], v)
	}
	value, err := nonIndexed.Unpack(l.Data)
	if err != nil {
		return i, err
	}
	for ix, v := range value {
		i.AddField(valueNames[ix], v)
	}

	return i, nil
}

func query(ceth *Ceth, fs encoding.FunctionSignature, data []byte) error {

	ctx := context.Background()

	account, contract, client, err := ceth.AccountContractClient()
	if err != nil {
		return err
	}
	address := contract.GetAddress()
	msg := ethereum.CallMsg{
		From: account.Address(),
		To:   &address,
		Data: data,
	}
	log.Debug().Hex("data", data).Hex("to", address.Bytes()).Hex("from", account.Address().Bytes()).Msg("CallContract")
	res, err := client.Client.CallContract(ctx, msg, nil)
	if err != nil {
		return errors.Wrap(err, "CallContract is failed")
	}
	if fs.Outputs != nil && len(fs.Outputs) > 0 {
		returned, err := fs.Outputs.Unpack(res)
		if err != nil {
			return err
		}

		stringResults := make([]string, 0)
		for _, r := range returned {
			switch v := r.(type) {
			case [32]uint8:
				stringResults = append(stringResults, hex.EncodeToString(v[:]))
			default:
				stringResults = append(stringResults, fmt.Sprintf("%s", r))

			}

		}
		fmt.Println(strings.Join(stringResults, ","))
	} else {
		fmt.Println(hex.EncodeToString(res))
	}
	return nil
}

func call(ceth *Ceth, value *big.Int, data []byte) error {
	ctx := context.Background()
	account, contract, client, err := ceth.AccountContractClient()
	if err != nil {
		return err
	}

	to := contract.GetAddress()
	tx, err := client.SubmitTx(ctx, account, &to, chain.WithData{Data: data}, chain.WithValue{Value: value})
	if err != nil {
		return err
	}

	chainClient, err := ceth.GetChainClient()
	if err != nil {
		return err
	}
	for i := 0; i < 30; i++ {
		info, err := chainClient.GetTransaction(ctx, tx)
		if err == nil {
			return PrintItem(info, "console")
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Submitted transaction couldn't been retrieved")

}

func deploy(ceth *Ceth, alias *string, value *string, contractFile string, constructorArgs []byte) error {
	content, err := ioutil.ReadFile(contractFile)
	if err != nil {
		return errors.Wrap(err, "Couldn't read file "+contractFile)
	}
	code := strings.TrimSpace(string(content))
	code = strings.TrimPrefix(code, "0x")

	codeData, err := encoding.HexToBytes(code)
	if err != nil {
		return errors.Wrapf(err, "Code %s is not proper HEX formatted", contractFile)
	}

	codeData = append(codeData, constructorArgs...)
	ctx := context.Background()

	account, client, err := ceth.AccountClient()
	if err != nil {
		return err
	}
	v := big.NewInt(0)
	if value != nil && *value != "" {
		var ok bool
		v, ok = v.SetString(*value, 10)
		if !ok {
			return fmt.Errorf("invald number")
		}
	}
	txHash, err := client.SubmitTx(ctx, account, nil, chain.WithData{Data: codeData}, chain.WithValue{Value: v})
	if err != nil {
		return err
	}

	var receipt *ethtypes.Receipt
	for {
		receipt, err = client.Client.TransactionReceipt(ctx, txHash)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break

	}
	fmt.Println("Contract: " + receipt.ContractAddress.Hex())
	fmt.Println("Transaction: " + receipt.TxHash.Hex())
	fmt.Println()
	if alias != nil && *alias != "" {
		chainID, err := client.Client.ChainID(ctx)
		if err != nil {
			return err
		}
		err = ceth.ContractRepo.AddContract(types.Contract{
			Name:    *alias,
			Address: receipt.ContractAddress.Hex(),
			ChainID: chainID.Int64(),
		})
		if err != nil {
			return err
		}
	}
	return nil

}
