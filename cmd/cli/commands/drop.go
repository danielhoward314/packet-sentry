package commands

import (
	"github.com/spf13/cobra"
)

// dropCmd is a subcommand that runs or rolls back goose migrations
var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Parent command for [db] commands for dropping a database.",
	Long:  "Parent command for [db] commands for dropping a database.",
}

func init() {
	rootCmd.AddCommand(dropCmd)
}
