package database

import (
	"database/sql"
	"fmt"
	"goflow/internal/testutils"
	"os"
	"path"
	"testing"
)

var databaseFile = path.Join(testutils.GetTestFolder(), "test.sqlite3")
var client *SQLClient

const testTable = "test"

var createTableQuery = fmt.Sprintf("create table %s(id integer, name string)", testTable)

func removeDBFile() {
	if _, err := os.Stat(databaseFile); err == nil {
		os.Remove(databaseFile)
	}
}

func TestMain(m *testing.M) {
	client = getTestSQLiteClient()
	removeDBFile()
	m.Run()
}

func getTestSQLiteClient() *SQLClient {
	return NewSQLiteClient(databaseFile)
}

func purgeDB() {
	rows, err := client.database.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	client.database.Begin()
	tables := make([]string, 0)
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			panic(err)
		}
		tables = append(tables, name)
	}
	for _, table := range tables {
		_, err = client.database.Exec(fmt.Sprintf("DROP TABLE %s", table))
		if err != nil {
			panic(err)
		}
	}
}

func TestNewDatabaseConnection(t *testing.T) {
	err := client.database.Ping()
	if err != nil {
		t.Error(err)
	}
}

func TestRunDatabaseQuery(t *testing.T) {
	defer purgeDB()
	err := client.Exec(createTableQuery)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateTable(t *testing.T) {
	defer purgeDB()
	client.createTable(table{
		name: "test",
		cols: []column{{"column1", "string"}, {"column2", "int"}},
	})
}

type resultType struct {
	id   int
	name string
}

type testRowResult struct {
	rows         *sql.Rows
	returnedRows []resultType
}

func (result testRowResult) ScanAppend() error {
	row := resultType{}
	err := result.rows.Scan(&row.id, &row.name)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func TestInsertIntoTable(t *testing.T) {
	defer purgeDB()
	_, err := client.database.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
	expectedID := 2
	expectedName := "yes"
	client.Insert(testTable, []string{"id", "name"}, []string{fmt.Sprint(expectedID), "'yes'"})
	rows, err := client.database.Query(fmt.Sprintf("SELECT * FROM %s", testTable))
	if err != nil {
		panic(err)
	}
	returnedRows := make([]resultType, 0)
	result := testRowResult{rows, returnedRows}
	putNRowValues(rows, result, 0)
	firstRow := returnedRows[0]
	if firstRow.name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, firstRow.name)
	}
	if firstRow.id != expectedID {
		t.Errorf("Expected id %d, got %d", expectedID, firstRow.id)
	}
}
