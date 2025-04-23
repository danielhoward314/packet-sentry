package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// migrateUpCmd is a subcommand that runs goose migrations
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Runs the sql migrations",
	Long:  "Runs the sql migrations",
	Run:   migrateUp,
}

func migrateUp(cobraCmd *cobra.Command, args []string) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	user := os.Getenv("POSTGRES_USER")
	applicationDB := os.Getenv("POSTGRES_APPLICATION_DATABASE")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		applicationDB,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
}
