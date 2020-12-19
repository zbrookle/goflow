package dagrun

import (
	dagtable "goflow/internal/dag/sql/dag"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"path"
	"testing"
)

var sqlClient *database.SQLClient
var tableClient *TableClient

var databaseFile = path.Join(testutils.GetTestFolder(), "test.sqlite3")
var testDagRow = dagtable.NewRow(0, "dag_num_1", "default", "v1", "/my/path", "json")

func setUpDagTable() {
	dagTableClient := dagtable.NewTableClient(sqlClient)
	dagTableClient.CreateTable()
	dagTableClient.UpsertDag(testDagRow)
}

func TestMain(m *testing.M) {
	sqlClient = database.NewSQLiteClient(databaseFile)
	tableClient = NewTableClient(sqlClient)
	m.Run()
}

func TestCreateDagRunTable(t *testing.T) {
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

func TestGetLastNDagRuns(t *testing.T) {

}

func TestUpsertDagRun(t *testing.T) {

}
