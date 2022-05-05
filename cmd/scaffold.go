package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/scaffold"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Scaffold a project for specified language",
}

func init() {
	scaffoldCmd.AddCommand(scaffold.GoCmd)
	dbCmd.AddCommand(scaffoldCmd)
}
