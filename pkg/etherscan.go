package cethacea

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func init() {

	tokenCmd := cobra.Command{
		Use:   "etherscan",
		Short: "Commands utilizes etherscan API",
	}
	RootCmd.AddCommand(&tokenCmd)
	{
		cmd := cobra.Command{
			Use:   "source",
			Short: "Download source of a contract",
			Args:  cobra.MaximumNArgs(1),
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ceth, err := NewCethContext(&Settings)
			if err != nil {
				return err
			}
			dest := ""
			if len(args) > 0 {
				dest = args[0]
			}
			return downloadSource(ceth, dest)
		}
		tokenCmd.AddCommand(&cmd)
	}

}

func downloadSource(ceth *Ceth, dest string) error {
	contract, err := ceth.GetCurrentContract()
	if err != nil {
		return err
	}

	url := "https://api-goerli.etherscan.io/api?module=contract&action=getsourcecode&address=" + contract.Address
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var parser fastjson.Parser

	parsed, err := parser.ParseBytes(body)
	if err != nil {
		return err
	}

	status := string(parsed.GetStringBytes("status"))
	if status != "1" {
		return errors.New("Unexpected status code: " + status)
	}

	result := parsed.GetArray("result")[0]

	sourceCode := result.GetStringBytes("SourceCode")
	sourceCode = sourceCode[:len(sourceCode)-1]
	sourceCode = sourceCode[1:]

	sources, err := parser.ParseBytes(sourceCode)
	if err != nil {
		return err
	}

	sources.GetObject("sources").Visit(func(key []byte, v *fastjson.Value) {
		name := string(key)
		name = strings.ReplaceAll(name, "/", "_")
		code := v.GetStringBytes("content")
		_ = os.Mkdir(contract.GetAddress().String(), 0755)
		err = ioutil.WriteFile(path.Join(contract.GetAddress().String(), name), code, 0644)
	})

	return nil
}
