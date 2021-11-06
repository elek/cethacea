package cethacea

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Settings CethSettings

var RootCmd = cobra.Command{
	Use: "cethacea",
}

func init() {
	RootCmd.PersistentFlags().StringVar(&Settings.Account, "account", "", "Account to be used for operations.")
	RootCmd.PersistentFlags().StringVar(&Settings.Contract, "contract", "", "Override the contract address to be used.")
	RootCmd.PersistentFlags().StringVar(&Settings.Chain, "chain", "", "RPC url of the used chain.")
	RootCmd.PersistentFlags().StringVar(&Settings.Abi, "abi", "", "Override the ABI of the current contract.")
	RootCmd.PersistentFlags().StringVar(&Settings.Format, "format", "console", "Format of the output (when applicable): console (default), json")
	RootCmd.PersistentFlags().BoolVar(&Settings.All, "all", false, "Use all chains/contracts/accounts including predefined and the ones from the other chains")
	RootCmd.PersistentFlags().BoolVar(&Settings.Debug, "debug", false, "Turn on debug level logging")
	_ = viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account"))
	_ = viper.BindPFlag("contract", RootCmd.PersistentFlags().Lookup("contract"))
	_ = viper.BindPFlag("chain", RootCmd.PersistentFlags().Lookup("chain"))
	_ = viper.BindPFlag("abi", RootCmd.PersistentFlags().Lookup("abi"))

}
