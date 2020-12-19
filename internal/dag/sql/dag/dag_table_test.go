package dag

import (
	"fmt"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"os"
	"path"
	"testing"
	"time"
)

var sqlClient *database.SQLClient
var tableClient *TableClient

var databaseFile = path.Join(testutils.GetTestFolder(), "test.sqlite3")

func TestMain(m *testing.M) {
	if err, _ := os.Stat(databaseFile); err == nil {
		os.Remove(databaseFile)
	}
	sqlClient = database.NewSQLiteClient(databaseFile)
	// defer database.PurgeDB(sqlClient)
	tableClient = NewTableClient(sqlClient)
	m.Run()
}

func TestCreateDagTable(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	tableClient.CreateTable()
	found := false
	for _, table := range sqlClient.Tables() {
		if table == TableName {
			found = true
		}
	}
	if !found {
		t.Errorf("Did not find table %s in tables", TableName)
	}
}

func createTestTable() {
	tableClient.sqlClient.CreateTable(tableClient.tableDef)
}

func TestIsDagInDagTable(t *testing.T) {
	defer database.PurgeDB(sqlClient)

	createTestTable()

	const dagName = "test"
	const namespace = "default"
	errMessageSuffix := fmt.Sprintf("be present for dag '%s' in namespace '%s'", dagName, namespace)
	if tableClient.IsDagPresent(dagName, namespace) {
		t.Errorf("Record should not " + errMessageSuffix)
	}
	sqlClient.Insert(
		TableName,
		Row{
			ID:              0,
			Name:            dagName,
			Namespace:       namespace,
			Version:         "0.0.1",
			FilePath:        "path",
			FileFormat:      "json",
			CreatedDate:     time.Now(),
			LastUpdatedDate: time.Now(),
		}.columnar(),
	)
	if !tableClient.IsDagPresent(dagName, namespace) {
		t.Errorf("Record should " + errMessageSuffix)
	}
}

func getTestRows() []Row {
	result := dagRowResult{}
	tableClient.sqlClient.QueryIntoResults(&result, "SELECT * FROM "+TableName)
	return result.returnedRows
}

func TestUpsertDagTable(t *testing.T) {
	defer database.PurgeDB(sqlClient)

	createTestTable()

	expectedRow := Row{
		ID:              0,
		Name:            "test",
		Namespace:       "default",
		Version:         "0.1.0",
		FilePath:        "path",
		FileFormat:      "json",
		CreatedDate:     time.Time{},
		LastUpdatedDate: time.Time{},
	}

	tableClient.UpsertDag(expectedRow)

	rows := getTestRows()
	rowCount := len(rows)
	if rowCount != 1 {
		t.Errorf("Expected only 1 row, found %d", rowCount)
	}

	if rows[0] != expectedRow {
		t.Errorf(
			"Expected %s, got %s",
			expectedRow,
			rows[0],
		)
	}

	expectedRow.Version = "0.2.0"
	tableClient.UpsertDag(expectedRow)

	rows = getTestRows()
	if rowCount != 1 {
		t.Errorf("Expected only 1 row, found %d", rowCount)
	}
	if rows[0] != expectedRow {
		t.Errorf(
			"Expected %s, got %s",
			expectedRow,
			rows[0],
		)
	}
}
