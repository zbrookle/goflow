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

var idColumn = Column{idName, Int{}}
var nameColumn = Column{nameName, String{}}

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
	name2Column := Column{"name2Name", String{}}
	tables := []Table{
		{Name: testTable, Cols: []Column{idColumn, nameColumn}},
		{
			Name:          testTable,
			Cols:          []Column{idColumn, nameColumn},
			PrimaryKeyCol: Column{idName, Int{}},
		},
		{
			Name:       testTable,
			Cols:       []Column{idColumn, nameColumn, name2Column},
			UniqueCols: []Column{nameColumn, name2Column},
		},
		{
			Name:          testTable,
			Cols:          []Column{idColumn, nameColumn, name2Column},
			PrimaryKeyCol: Column{idName, Int{}},
			UniqueCols:    []Column{nameColumn, name2Column},
		},
	}

	for _, table := range tables {
		func() {
			defer PurgeDB(client)
			client.CreateTable(table)

			rows, err := client.database.Query(
				fmt.Sprintf(
					"SELECT name FROM sqlite_master WHERE type = 'table' and name = '%s'",
					testTable,
				),
			)
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
			if len(names) != 1 {
				t.Error("Expected one table from query")
			}
		}()

	}
}

func TestCreateTablesForeignKey(t *testing.T) {
	defer PurgeDB(client)
	t1 := Table{
		Name:          testTable,
		Cols:          []Column{idColumn, nameColumn},
		PrimaryKeyCol: Column{idName, Int{}},
	}

	t2IdColumn := Column{"t1Id", Int{}}
	t2 := Table{
		Name: "table2",
		Cols: []Column{
			t2IdColumn,
			{
				"otherColumn",
				String{},
			},
		},
		ForeignKeys: []KeyReference{{
			Key:      t2IdColumn,
			RefTable: t1.Name,
			RefCol:   idColumn,
		},
		}}
	client.CreateTable(t1)
	client.CreateTable(t2)
}

type resultType struct {
	id   int
	name string
}

type testQueryResult struct {
	rows         *sql.Rows
	returnedRows []resultType
}

func (result *testQueryResult) ScanAppend(rows *sql.Rows) error {
	row := resultType{}
	err := rows.Scan(&row.id, &row.name)
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

func getRowsFromTestTable() []resultType {
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
	return returnedRows
}

func TestInsertIntoTable(t *testing.T) {
	defer PurgeDB(client)
	_, err := client.database.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
	client.Insert(
		testTable,
		[]ColumnWithValue{
			{Column{idName, Int{expectedID}}},
			{Column{nameName, String{expectedName}}},
		},
	)

	returnedRows := getRowsFromTestTable()
	firstRow := returnedRows[0]
	if firstRow.name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, firstRow.name)
	}
	if firstRow.id != expectedID {
		t.Errorf("Expected id %d, got %d", expectedID, firstRow.id)
	}
}

func setUpTableAndInsertOneRow() {
	// Set up table
	_, err := client.database.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
	_, err = client.database.Exec(insertionQuery)
	if err != nil {
		panic(err)
	}
}

func TestQueryRowsIntoResult(t *testing.T) {
	defer PurgeDB(client)

	setUpTableAndInsertOneRow()

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

func TestUpdateTable(t *testing.T) {
	defer PurgeDB(client)

	setUpTableAndInsertOneRow()

	const newName = "no"
	client.Update(
		testTable,
		[]ColumnWithValue{{Column{nameName, String{newName}}}},
		[]ColumnWithValue{{Column{idName, Int{expectedID}}}},
	)
	rows := getRowsFromTestTable()
	if rows[0].name != newName {
		t.Errorf("Column %s was not updated from %s to %s", nameName, expectedName, newName)
	}
}
