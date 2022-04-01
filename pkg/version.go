package cethacea

import (
	"debug/buildinfo"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func init() {

	version := cobra.Command{
		Use:   "version",
		Short: "Show embedded version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			ex, err := os.Executable()
			if err != nil {
				return err
			}
			info, err := buildinfo.ReadFile(ex)
			if err != nil {
				return err
			}
			for _, v := range info.Settings {
				if strings.HasPrefix(v.Key, "vcs.") {
					fmt.Printf("%s: %s\n", v.Key, v.Value)
				}
			}
			return nil
		},
	}
	RootCmd.AddCommand(&version)
}
