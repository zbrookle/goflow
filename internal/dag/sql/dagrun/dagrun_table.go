package dagrun

import (
	"fmt"
	"goflow/internal/database"
	"time"
)

const tableName = "dagrun"

// TableClient is a struct that interacts with the DAG table
type TableClient struct {
	sqlClient *database.SQLClient
	tableDef  database.Table
}

// NewTableClient returns a new table client
func NewTableClient(sqlClient *database.SQLClient) *TableClient {
	return &TableClient{sqlClient, database.Table{Name: tableName,
		Cols: []database.Column{
			{Name: "dag_id", DType: database.Int{}},
			{Name: "status", DType: database.String{}},
			{Name: "execution_date", DType: database.TimeStamp{}},
			{Name: "start_date", DType: database.TimeStamp{}},
			{Name: "end_date", DType: database.TimeStamp{}},
			{Name: "last_updated_date", DType: database.TimeStamp{}},
		},
	}}
}

// CreateTable creates the table for storing DAG related information
func (client *TableClient) CreateTable() {
	client.sqlClient.CreateTable(client.tableDef)
}

// GetLastNRunsForDagID retrieves the rows for a given dag id
func (client *TableClient) GetLastNRunsForDagID(dagID int, n int) dagRowResult {
	result := newRowResult(n)
	client.sqlClient.QueryIntoResults(
		&result,
		fmt.Sprintf(
			"SELECT * FROM dagrun WHERE dag_id = %d ORDER BY last_updated_date desc",
			dagID,
		),
	)
	return result
}

func (client *TableClient) selectSpecificDagRun(dagID int, executionDate time.Time) dagRowResult {
	result := newRowResult(1)
	client.sqlClient.QueryIntoResults(
		&result,
		fmt.Sprintf(
			"SELECT * FROM dagrun WHERE dag_id = %d AND executionDate = %s ORDER BY last_updated_date desc",
			dagID,
			executionDate,
		),
	)
	return result
}

func (client *TableClient) isDagRunPresent(dagID int, executionDate time.Time) bool {
	rows := client.selectSpecificDagRun(dagID, executionDate)
	return len(rows.returnedRows) == 1
}

// UpsertDagRun inserts or updates the dag run
func (client *TableClient) UpsertDagRun(dagRunRow Row) {
	switch client.isDagRunPresent(dagRunRow.dagID, dagRunRow.executionDate) {
	case false:
		client.sqlClient.Insert(tableName, dagRunRow.columnar())
	default:
		client.sqlClient.Update(tableName,
			[]database.ColumnWithValue{
				{
					Column: database.Column{Name: statusName, DType: database.String{}},
					Value:  dagRunRow.status,
				},
			},
			[]database.ColumnWithValue{
				{
					Column: database.Column{Name: dagIDName, DType: database.String{}},
					Value:  fmt.Sprint(dagRunRow.dagID),
				},
				{
					Column: database.Column{Name: executionDateName, DType: database.String{}},
					Value:  dagRunRow.executionDate.String(),
				},
			})
	}
}
