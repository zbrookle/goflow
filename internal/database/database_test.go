package database

import (
	"fmt"
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

func TestCreateTable(t *testing.T) {
	defer purgeDB()
	client.createTable(table{
		name: "test",
		cols: []column{{"column1", "string"}, {"column2", "int"}},
	})
}

func TestRunDatabaseQuery(t *testing.T) {
	defer purgeDB()
	err := client.Exec(`create table test(id integer, name string)`)
	if err != nil {
		t.Error(err)
	}
}
