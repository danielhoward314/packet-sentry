package commands

import (
	"github.com/spf13/cobra"
)

// createCmd is a subcommand that creates a db or a migration file
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Parent command for [db|migration] commands for creating a database or a migration file",
	Long:  "Parent command for [db|migration] commands for creating a database or a migration file",
}

func init() {
	rootCmd.AddCommand(createCmd)
}
