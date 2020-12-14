package dag

import (
	"fmt"
	"goflow/internal/database"
)

const tableName = "dags"

// TableClient is a struct that interacts with the DAG table
type TableClient struct {
	sqlClient *database.SQLClient
	tableDef  database.Table
}

// NewTableClient returns a new table client
func NewTableClient(sqlClient *database.SQLClient) *TableClient {
	return &TableClient{sqlClient, database.Table{Name: tableName,
		Cols: []database.Column{
			{Name: "id", DType: database.Int{}},
			{Name: "name", DType: database.String{}},
			{Name: "namespace", DType: database.String{}},
			{Name: "version", DType: database.String{}},
			{Name: "file_path", DType: database.String{}},
			{Name: "file_format", DType: database.String{}},
			{Name: "created_date", DType: database.TimeStamp{}},
			{Name: "last_updated_date", DType: database.TimeStamp{}},
		},
	}}
}

// CreateTable creates the table for storing DAG related information
func (client *TableClient) CreateTable() {
	client.sqlClient.CreateTable(client.tableDef)
}

func (client *TableClient) selectSpecificDag(name, namespace string) []Row {
	result := newRowResult(1)
	client.sqlClient.QueryIntoResults(
		&result,
		fmt.Sprintf(
			"SELECT * FROM %s WHERE name = '%s' and namespace = '%s'",
			tableName,
			name,
			namespace,
		),
	)
	rowCount := len(result.returnedRows)
	if rowCount > 1 {
		panic(fmt.Sprintf("should be 1 or fewer dags, found %d", rowCount))
	}
	return result.returnedRows
}

// GetDagRecord returns a record for a dag and ok if the record exists
func (client *TableClient) GetDagRecord(name, namespace string) Row {
	rows := client.selectSpecificDag(name, namespace)
	return rows[0]
}

// IsDagPresent returns true if the dag is in the dags table
func (client *TableClient) IsDagPresent(name, namespace string) bool {
	rows := client.selectSpecificDag(name, namespace)
	return len(rows) == 1
}

// UpsertDag inserts a new dag if it does not exist or updates
// an existing dag record
func UpsertDag() {

}
