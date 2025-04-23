package commands

import (
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// createMigrationCmd is a subcommand to create a migration .sql file with the current timestamp
var createMigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Creates a .sql migration file with name <timestamp>_<name>.sql where the <name> comes from `create migration <name>`",
	Long:  "Creates a .sql migration file with name <timestamp>_<name>.sql where the <name> comes from `create migration <name>`",
	Run:   createMigration,
}

func createMigration(cobraCmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatal("must provide name for migration file")
	}
	migrationName := args[0]

	goose.SetBaseFS(embedMigrations)
	err := goose.Create(nil, "./migrations", migrationName, "sql")
	if err != nil {
		log.Fatalf("Error creating migration file: %s", err.Error())
	}
}

func init() {
	createCmd.AddCommand(createMigrationCmd)
}
