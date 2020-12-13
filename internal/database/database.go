package database

import (
	"database/sql"
	"fmt"
	"strings"

	sqlite "github.com/mattn/go-sqlite3"
)

const sqliteDriver = "sqlite"
const unlimitedCapacity = 0

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
	Capacity() int
}

// PutNRowValues puts the first RowResult.Length rows into row result
func PutNRowValues(result RowResult) {
	defer result.Rows().Close()
	i := 0
	for result.Rows().Next() {
		fmt.Println("Here!!!")
		if i == result.Capacity() && result.Capacity() != unlimitedCapacity {
			break
		}
		err := result.ScanAppend()
		if err != nil {
			panic(err)
		}
		i++
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

// CreateTable creates a table in the database for the given SQLClient
func (client *SQLClient) CreateTable(t Table) {
	query := t.createQuery()
	err := client.Exec(query)
	if err != nil {
		panic(fmt.Sprintf("error '%s' occurred for query '%s'", err.Error(), query))
	}
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
