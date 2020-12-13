package dag

import (
	"fmt"
	"goflow/internal/database"
)

const tableName = "dags"

// TableClient is a struct that interacts with the DAG table
type TableClient struct {
	sqlClient *database.SQLClient
}

// CreateTable creates the table for storing DAG related information
func (client *TableClient) CreateTable() {
	table := database.Table{Name: tableName,
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
	}
	client.sqlClient.CreateTable(table)
}

// GetDagRecord returns a record for a dag and ok if the record exists
func (client *TableClient) GetDagRecord(name, namespace string) Row {
	result := newRowResult(1)
	client.sqlClient.QueryIntoResults(&result, fmt.Sprintf("SELECT * FROM %s", tableName))
	return result.returnedRows[0]
}

// UpsertDag inserts a new dag if it does not exist or updates
// an existing dag record
func UpsertDag() {
}
