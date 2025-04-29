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
	Short: "Runs `CREATE DATABASE <name>;` with TEMPLATE template0 for clean collation. Also creates TimescaleDB database and extension.",
	Long:  "Runs `CREATE DATABASE <name> TEMPLATE template0 LC_COLLATE 'C' LC_CTYPE 'C' ENCODING 'UTF8';` for both the app database and the TimescaleDB database.",
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

	createAppDBSQL := fmt.Sprintf(`
		CREATE DATABASE %s
		WITH
		  TEMPLATE template0
		  ENCODING 'UTF8'
		  LC_COLLATE 'C'
		  LC_CTYPE 'C';
	`, pqQuoteIdentifier(applicationDB))

	_, err = db.Exec(createAppDBSQL)
	if err != nil {
		log.Fatal("Error creating application database:", err)
	}
	fmt.Printf("Database %s created successfully.\n", applicationDB)

	// Create TimescaleDB Database
	tsdbName := os.Getenv("TSDB_DATABASE")
	if tsdbName == "" {
		log.Fatal("Error creating TimescaleDB database: empty TSDB_DATABASE name")
	}

	createTSDBSQL := fmt.Sprintf(`
		CREATE DATABASE %s
		WITH
		  TEMPLATE template0
		  ENCODING 'UTF8'
		  LC_COLLATE 'C'
		  LC_CTYPE 'C';
	`, pqQuoteIdentifier(tsdbName))

	_, err = db.Exec(createTSDBSQL)
	if err != nil {
		log.Fatal("Error creating TimescaleDB database:", err)
	}
	fmt.Printf("TimescaleDB database %s created successfully.\n", tsdbName)
}

func init() {
	createCmd.AddCommand(createDBCmd)
}

// pqQuoteIdentifier safely quotes database identifiers (like db names)
func pqQuoteIdentifier(identifier string) string {
	return `"` + identifier + `"`
}
