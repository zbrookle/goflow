package metrics

import (
	"fmt"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"path"
	"testing"
	"time"
)

var sqlClient *database.SQLClient
var tableClient *TableClient

var databaseFile = path.Join(testutils.GetTestFolder(), "test.sqlite3")

const testName = "test"

func TestMain(m *testing.M) {
	sqlClient = database.NewSQLiteClient(databaseFile)
	tableClient = NewTableClient(sqlClient)
	m.Run()
}

func TestCreateMetricsTable(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	tableClient.CreateTable()
	found := false
	for _, table := range sqlClient.Tables() {
		if table == tableName {
			found = true
		}
	}
	if !found {
		t.Errorf("Did not find table %s in tables", tableName)
	}
}

func getTime(i int) time.Time {
	timeStr := fmt.Sprintf("2019-01-%02d", i+1)
	executionTime, err := time.Parse("2006-01-02", timeStr)
	if err != nil {
		panic(err)
	}
	return executionTime
}

func insertMetrics(n int) []Row {
	insertedRows := make([]Row, 0, n)
	for i := 0; i < n; i++ {
		timeStr := fmt.Sprintf("2019-01-%02d", i+1)
		executionTime := getTime(i)
		newRow := NewRow(i, testName, testName+timeStr, 3000, 4000, executionTime)
		insertedRows = append(insertedRows, newRow)
		sqlClient.Insert(tableName, newRow.columnar())
	}
	fmt.Println(insertedRows)
	return insertedRows
}

func setUpTestTable() {
	sqlClient.CreateTable(tableClient.tableDef)
}

func TestGetMetricsForDAG(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	setUpTestTable()

	const insertedDays = 5
	expectedRows := insertMetrics(insertedDays)

	foundRows := tableClient.GetMetricsForDag(testName, getTime(0), getTime(insertedDays))
	fmt.Println(foundRows)

	length := len(foundRows)
	if length != insertedDays {
		t.Errorf("Expected %d rows, found %d", insertedDays, length)
	}
	for i, row := range foundRows {
		expectedRow := expectedRows[i]
		if row != expectedRow {
			t.Errorf("expected row %s, found row %s", expectedRow, row)
		}
	}

}
