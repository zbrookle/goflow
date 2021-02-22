package dag

import (
	"database/sql"
	"fmt"
	"goflow/internal/database"
)

const nameName = "name"
const namespaceName = "namespace"

// TableName is the name of the dag table
const TableName = "dags"

// TableClient is a struct that interacts with the DAG table
type TableClient struct {
	sqlClient *database.SQLClient
	tableDef  database.Table
}

// NewTableClient returns a new table client
func NewTableClient(sqlClient *database.SQLClient) *TableClient {
	return &TableClient{sqlClient, database.Table{Name: TableName,
		Cols: Row{}.columnar().Columns(),
		PrimaryKeyCol: database.Column{
			Name:  IDName,
			DType: database.Int{},
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
			TableName,
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

func scanInt(rows *sql.Rows, intPtr *int) {
	if rows.Next() {
		rows.Scan(intPtr)
		err := rows.Close()
		if err != nil {
			panic(err)
		}
		return
	}
	panic("no rows found!")
}

// DagCount returns the number of dags in the database
func (client *TableClient) DagCount() int {
	rows, err := client.sqlClient.Query(fmt.Sprintf("SELECT COUNT(*) as count FROM %s", TableName))
	if err != nil {
		panic(err)
	}
	var count int
	scanInt(rows, &count)
	return count
}

// GetMostRecentID returns a new id
func (client *TableClient) GetMostRecentID() int {
	if client.DagCount() == 0 {
		return 0
	}
	rows, err := client.sqlClient.Query(fmt.Sprintf("SELECT MAX(id) FROM %s", TableName))
	if err != nil {
		panic(err)
	}
	var rowNumber int
	scanInt(rows, &rowNumber)
	return rowNumber + 1
}

// UpdateDAGToggle updates the on/off status of the DAG in the DB
func (client *TableClient) UpdateDAGToggle(dagID int, newState bool) {
	client.sqlClient.Update(
		TableName,
		database.ColumnWithValueSlice{
			{Column: database.Column{Name: isOnName, DType: database.Bool{Val: newState}}},
		},
		database.ColumnWithValueSlice{
			{Column: database.Column{Name: IDName, DType: database.Int{Val: dagID}}},
		},
	)
}

// UpsertDAG inserts a new dag if it does not exist or updates
// an existing dag record
func (client *TableClient) UpsertDAG(dagRow Row) Row {
	dagPresent := client.IsDagPresent(dagRow.Name, dagRow.Namespace)
	switch dagPresent {
	case false:
		dagRow.ID = client.GetMostRecentID()
		client.sqlClient.Insert(TableName, dagRow.columnar())
	default:
		originalRow := client.GetDagRecord(dagRow.Name, dagRow.Namespace)
		dagRow.ID = originalRow.ID
		client.sqlClient.Update(
			TableName,
			dagRow.columnar(),
			[]database.ColumnWithValue{
				{
					Column: database.Column{
						Name:  nameName,
						DType: database.String{Val: dagRow.Name},
					},
				},
				{
					Column: database.Column{
						Name:  namespaceName,
						DType: database.String{Val: dagRow.Namespace},
					},
				},
			},
		)
	}
	return dagRow
}
