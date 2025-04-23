package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// createDBCmd is a subcommand to create a database
var createDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Runs `CREATE DATABASE <name>;` where <name> is sourced from env var POSTGRES_APPLICATION_DATABASE.",
	Long:  "Runs `CREATE DATABASE <name>;` where <name> is sourced from env var POSTGRES_APPLICATION_DATABASE.",
	Run:   createDB,
}

func createDB(cobraCmd *cobra.Command, args []string) {
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
		log.Fatal("Error creating database: empty application database name")
	}
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s", applicationDB)
	_, err = db.Exec(createDBSQL)
	if err != nil {
		log.Fatal("Error creating database:", err)
	}

	fmt.Printf("Database %s created successfully.\n", applicationDB)
}

func init() {
	createCmd.AddCommand(createDBCmd)
}
