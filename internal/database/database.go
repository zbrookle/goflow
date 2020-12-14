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

// PurgeDB removes all tables from the database
func PurgeDB(client *SQLClient) {
	tables := client.Tables()
	for _, table := range tables {
		_, err := client.database.Exec(fmt.Sprintf("DROP TABLE %s", table))
		if err != nil {
			panic(err)
		}
	}
}

// QueryResult is implemented to retrieve the result for a rows object
type QueryResult interface {
	ScanAppend() error
	Rows() *sql.Rows
	Capacity() int
	SetRows(*sql.Rows)
}

// PutNRowValues puts the first RowResult.Length rows into row result
func PutNRowValues(result QueryResult) {
	defer result.Rows().Close()
	i := 0
	for result.Rows().Next() {
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

// QueryIntoResults places query results into a structure of interface RowResult
func (client *SQLClient) QueryIntoResults(result QueryResult, queryString string) {
	rows, err := client.Query(queryString)
	if err != nil {
		panic(err)
	}
	result.SetRows(rows)
	PutNRowValues(result)
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
func (client *SQLClient) Insert(table string, columns []Column, values []string) {
	if len(columns) != len(values) {
		panic("columns and values must be the same length")
	}
	commaJoin := func(s []string) string { return strings.Join(s, ",") }
	columnNames := make([]string, 0, len(columns))
	valueStrings := make([]string, 0, len(values))
	for i, col := range columns {
		columnNames = append(columnNames, col.Name)
		valueStrings = append(valueStrings, col.DType.getValRep(values[i]))
	}
	err := client.Exec(
		fmt.Sprintf(
			"INSERT INTO %s(%s) VALUES(%s)",
			table,
			commaJoin(columnNames),
			commaJoin(valueStrings),
		),
	)
	if err != nil {
		panic(err)
	}
}

// Tables returns a list of the table names in the databse
func (client *SQLClient) Tables() []string {
	rows, err := client.database.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	fmt.Println(names)
	return names
}
