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
	defer database.PurgeDB(sqlClient)
	tableClient = NewTableClient(sqlClient)
	m.Run()
}

func TestCreateDagTable(t *testing.T) {
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

func TestIsDagInDagTable(t *testing.T) {
	defer database.PurgeDB(sqlClient)

	tableClient.CreateTable()

	const dagName = "test"
	const namespace = "default"
	errMessageSuffix := fmt.Sprintf("be present for dag '%s' in namespace '%s'", dagName, namespace)
	if tableClient.IsDagPresent(dagName, namespace) {
		t.Errorf("Record should not " + errMessageSuffix)
	}
	sqlClient.Insert(
		tableName,
		tableClient.tableDef.Cols,
		[]string{
			"0",
			dagName,
			namespace,
			"0.0.1",
			"path",
			"json",
			time.Now().String(),
			time.Now().String(),
		},
	)
	if !tableClient.IsDagPresent(dagName, namespace) {
		t.Errorf("Record should " + errMessageSuffix)
	}
}

func TestUpsertDagTable(t *testing.T) {

}
