package commands

import (
	"fmt"

	"github.com/go-fusion/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.MainVersion.StringValue)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
