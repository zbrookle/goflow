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
	ScanAppend(*sql.Rows) error
	Capacity() int
}

// queryErrorMessage returns the error message along with the associated query
func queryErrorMessage(query string, err error) string {
	return fmt.Sprintf("for query: '%s': %s", query, err.Error())
}

// PutNRowValues puts the first RowResult.Length rows into row result
func PutNRowValues(result QueryResult, rows *sql.Rows) {
	defer rows.Close()
	i := 0
	for rows.Next() {
		if i == result.Capacity() && result.Capacity() != unlimitedCapacity {
			break
		}
		err := result.ScanAppend(rows)
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
	PutNRowValues(result, rows)
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
		panic(queryErrorMessage(query, err))
	}
}

// Insert inserts rows into a given table in the database
func (client *SQLClient) Insert(table string, columns []ColumnWithValue) {
	commaJoin := func(s []string) string { return strings.Join(s, ",") }
	columnNames := make([]string, 0, len(columns))
	valueStrings := make([]string, 0, len(columns))
	for _, col := range columns {
		columnNames = append(columnNames, col.Name)
		valueStrings = append(valueStrings, col.ValRep())
	}
	query := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES(%s)",
		table,
		commaJoin(columnNames),
		commaJoin(valueStrings),
	)
	err := client.Exec(query)
	if err != nil {
		panic(queryErrorMessage(query, err))
	}
}

// Update updates a given table with row values, for a given condition
func (client *SQLClient) Update(table string, values, conditions ColumnWithValueSlice) {
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		table,
		values.Join(", "),
		conditions.Join(" AND "),
	)
	err := client.Exec(query)
	if err != nil {
		panic(queryErrorMessage(query, err))
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
	return names
}
