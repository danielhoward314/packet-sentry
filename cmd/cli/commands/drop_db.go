package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// dropDBCmd is a subcommand to drop a database
var dropDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Drops both the application and timescaledb databases.",
	Long:  "Drops the databases sourced from POSTGRES_APPLICATION_DATABASE and TSDB_DATABASE environment variables.",
	Run:   dropDB,
}

func dropDB(cobraCmd *cobra.Command, args []string) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	user := os.Getenv("POSTGRES_USER")
	mainDB := os.Getenv("POSTGRES_MAIN_DATABASE")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		mainDB,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	applicationDB := os.Getenv("POSTGRES_APPLICATION_DATABASE")
	if applicationDB == "" {
		log.Fatal("Error dropping database: empty application database name")
	}

	dropAppDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", pqQuoteIdentifier(applicationDB))
	_, err = db.Exec(dropAppDBSQL)
	if err != nil {
		log.Fatal("Error dropping application database:", err)
	}
	fmt.Printf("Database %s dropped successfully.\n", applicationDB)

	tsdbDatabase := os.Getenv("TSDB_DATABASE")
	if tsdbDatabase == "" {
		log.Fatal("Error dropping TimescaleDB database: empty TSDB_DATABASE name")
	}

	dropTSDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", pqQuoteIdentifier(tsdbDatabase))
	_, err = db.Exec(dropTSDBSQL)
	if err != nil {
		log.Fatal("Error dropping TimescaleDB database:", err)
	}
	fmt.Printf("TimescaleDB database %s dropped successfully.\n", tsdbDatabase)
}

func init() {
	dropCmd.AddCommand(dropDBCmd)
}
