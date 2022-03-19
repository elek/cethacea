package evm

import (
	"fmt"
	cethacea "github.com/elek/cethacea/pkg"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs/v2"
	"io/ioutil"
	"os/exec"
	"strings"
)

func init() {

	evmCmd := cobra.Command{
		Use:     "evm",
		Short:   "EVM related (contract compilation / decompilation / assembly) commands",
		Aliases: []string{"e"},
	}
	{
		cmd := cobra.Command{
			Use:     "disassembly",
			Short:   "Disassemble bytecode from file",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"d", "dasm"},
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return disasm(args[0])
		}
		evmCmd.AddCommand(&cmd)
	}
	{

		cmd := cobra.Command{
			Use:     "assembly",
			Short:   "Compile contract source to binary data",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"a", "asm", "compile"},
		}
		compiler := cmd.Flags().String("compiler", "internal", "Name of the compiler to use ('internal' or 'yulc')")
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			switch *compiler {
			case "yulc":
				return yulasm(args[0])
			case "internal":
				return asm(args[0])
			default:
				return errs.Errorf("No such compiler %s. Use yulc or internal.", *compiler)
			}
			return disasm(args[0])
		}
		evmCmd.AddCommand(&cmd)
	}
	cethacea.RootCmd.AddCommand(&evmCmd)
}

func yulasm(s string) error {
	c := exec.Command("solc", "--strict-assembly", s)
	output, err := c.Output()
	if err != nil {
		return errs.Wrap(err)
	}
	byteCode := ""
	codePart := false
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			codePart = false
		}
		if codePart {
			byteCode = byteCode + line
		}
		if line == "Binary representation:" {
			codePart = true
		}
	}
	fmt.Println(byteCode)
	err = ioutil.WriteFile(s+".bin", []byte(byteCode), 0644)
	if err != nil {
		return nil
	}
	return nil
}
