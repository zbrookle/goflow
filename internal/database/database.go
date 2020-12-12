package database

// What to store in database?
// Table for dags
// Table for dagruns

import (
	"database/sql"

	sqlite "github.com/mattn/go-sqlite3"
)

const sqliteDriver = "sqlite"

func init() {
	sql.Register(sqliteDriver, &sqlite.SQLiteDriver{})
}

// SQLClient uses a shared database connection to retrieves and store information about application state
type SQLClient struct {
	database *sql.DB
}

// NewSQLiteClient returns a new SQLlite Client
func NewSQLiteClient(dsn string) *SQLClient {
	db, err := sql.Open(sqliteDriver, dsn)
	if err != nil {
		panic(err)
	}
	return &SQLClient{
		database: db,
	}
}

// Query runs a database query and returns the rows
func (client *SQLClient) Query(queryString string) (*sql.Rows, error) {
	return client.database.Query(queryString)
}

// Exec runs a database query without returning rows
func (client *SQLClient) Exec(queryString string) error {
	_, err := client.database.Exec(queryString)
	return err
}

func (client *SQLClient) createTable(t table) error {
	return client.Exec(t.createQuery())
}

// SetupDatabase creates the database and necessary tables for the application
func (client *SQLClient) SetupDatabase() {
	err := client.database.Ping()
	if err != nil {
		panic(err)
	}
	err = client.createTable(table{
		name: "dags",
		cols: make([]column, 0),
	})
}
