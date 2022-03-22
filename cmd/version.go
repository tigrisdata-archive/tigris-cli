package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigrisdb-cli/util"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show tigrisdb-cli version",
	Run: func(cmd *cobra.Command, args []string) {
		util.Stdout("tigrisdb-cli version %s\n", util.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
