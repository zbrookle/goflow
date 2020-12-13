package database

import (
	"database/sql"
	"fmt"
	"strings"

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

// RowResult is implemented to retrieve the result for a rows object
type RowResult interface {
	ScanAppend() error
	Rows() *sql.Rows
}

// PutNRowValues returns
func PutNRowValues(result RowResult, n int) {
	defer result.Rows().Close()
	for result.Rows().Next() {
		err := result.ScanAppend()
		if err != nil {
			panic(err)
		}
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

func (client *SQLClient) createTable(t table) {
	query := t.createQuery()
	err := client.Exec(query)
	if err != nil {
		panic(fmt.Sprintf("error '%s' occurred for query '%s'", err.Error(), query))
	}
}

// SetupDatabase creates the database and necessary tables for the application
func (client *SQLClient) SetupDatabase() {
	err := client.database.Ping()
	if err != nil {
		panic(err)
	}
	client.createTable(table{
		name: "dags",
		cols: make([]column, 0),
	})
}

// Insert inserts rows into a given table in the database
func (client *SQLClient) Insert(table string, columns, values []string) {
	commaJoin := func(s []string) string { return strings.Join(s, ",") }
	err := client.Exec(
		fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, commaJoin(columns), commaJoin(values)),
	)
	if err != nil {
		panic(err)
	}
}
