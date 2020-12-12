package database

import (
	"goflow/internal/logs"
	"goflow/internal/testutils"
	"os"
	"path"
	"testing"
)

var databaseFile = path.Join(testutils.GetTestFolder(), "test.sqlite3")
var client *SQLClient

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
	if err != nil {
		panic(err)
	}
	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	logs.InfoLogger.Println(columns)
}

func TestNewDatabaseConnection(t *testing.T) {
	err := client.database.Ping()
	if err != nil {
		t.Error(err)
	}
}

func TestCreateTable(t *testing.T) {
	defer purgeDB()
	err := client.createTable(table{
		name: "dags",
		cols: []column{{"column1", "string"}, {"column2", "int"}},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestRunDatabaseQuery(t *testing.T) {
	defer purgeDB()
	err := client.Exec(`create table dags(id integer, name string)`)
	if err != nil {
		t.Error(err)
	}
}
