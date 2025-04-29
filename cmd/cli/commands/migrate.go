package commands

import (
	"embed"

	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

//go:embed migrations_timescale/*.sql
var embedTimescaleMigrations embed.FS

const (
	dirMigrations                   = "migrations"
	dirMigrationsTimescale          = "migrations_timescale"
	embeddedMigrationsPath          = "./migrations"
	embeddedTimescaleMigrationsPath = "./migrations_timescale"
)

// migrateCmd is a subcommand that runs or rolls back goose migrations
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Parent command for [up|down] commands for running or rolling back goose migrations",
	Long:  "Parent command for [up|down] commands for running or rolling back goose migrations",
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
