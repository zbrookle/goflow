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
const expectedID = 2
const expectedName = "yes"
const idName = "id"
const nameName = "name"

var createTableQuery = fmt.Sprintf("CREATE TABLE %s(%s INTEGER, %s STRING)", testTable, idName, nameName)

var insertionQuery = fmt.Sprintf("INSERT INTO %s(%s, %s) VALUES(%d, '%s')", testTable, idName, nameName, expectedID, expectedName)

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

func TestNewDatabaseConnection(t *testing.T) {
	err := client.database.Ping()
	if err != nil {
		t.Error(err)
	}
}

func TestRunDatabaseQuery(t *testing.T) {
	defer PurgeDB(client)
	err := client.Exec(createTableQuery)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateTable(t *testing.T) {
	defer PurgeDB(client)
	client.CreateTable(Table{
		Name: "test",
		Cols: []Column{{"column1", String{}}, {"column2", Int{}}},
	})
}

type resultType struct {
	id   int
	name string
}

type testQueryResult struct {
	rows         *sql.Rows
	returnedRows []resultType
}

func (result *testQueryResult) ScanAppend() error {
	row := resultType{}
	err := result.rows.Scan(&row.id, &row.name)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *testQueryResult) Rows() *sql.Rows {
	return result.rows
}

func (result *testQueryResult) Capacity() int {
	return cap(result.returnedRows)
}

func (result *testQueryResult) SetRows(rows *sql.Rows) {
	result.rows = rows
}

func TestInsertIntoTable(t *testing.T) {
	defer PurgeDB(client)
	_, err := client.database.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
	client.Insert(
		testTable,
		[]Column{{idName, Int{}}, {nameName, String{}}},
		[]string{fmt.Sprint(expectedID), expectedName},
	)
	rows, err := client.database.Query(fmt.Sprintf("SELECT * FROM %s", testTable))
	if err != nil {
		panic(err)
	}

	// Retrieve rows
	returnedRows := make([]resultType, 0, 1)
	for rows.Next() {
		result := resultType{}
		rows.Scan(&result.id, &result.name)
		returnedRows = append(returnedRows, result)
	}
	firstRow := returnedRows[0]
	if firstRow.name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, firstRow.name)
	}
	if firstRow.id != expectedID {
		t.Errorf("Expected id %d, got %d", expectedID, firstRow.id)
	}
}

func TestQueryRowsIntoResult(t *testing.T) {
	defer PurgeDB(client)

	// Set up table
	_, err := client.database.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
	_, err = client.database.Exec(insertionQuery)
	if err != nil {
		panic(err)
	}

	returnedRows := make([]resultType, 0, 1)
	result := testQueryResult{returnedRows: returnedRows}
	client.QueryIntoResults(&result, fmt.Sprintf("SELECT * FROM %s", testTable))
	firstRow := result.returnedRows[0]
	if firstRow.name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, firstRow.name)
	}
	if firstRow.id != expectedID {
		t.Errorf("Expected id %d, got %d", expectedID, firstRow.id)
	}
}
