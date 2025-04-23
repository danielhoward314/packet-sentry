package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// migrateDownCmd is a subcommand that rolls back goose migrations
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rolls back the sql migrations",
	Long:  "Rolls back the sql migrations",
	Run:   migrateDown,
}

func migrateDown(cobraCmd *cobra.Command, args []string) {
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

	if err := goose.Down(db, "migrations"); err != nil {
		panic(err)
	}
}

func init() {
	migrateCmd.AddCommand(migrateDownCmd)
}
