package metrics

import (
	"fmt"
	"goflow/internal/database"
	"goflow/internal/dateutils"
	"time"
)

const tableName = "metrics"

// TableClient is a struct that interacts with the DAG table
type TableClient struct {
	sqlClient *database.SQLClient
	tableDef  database.Table
}

// NewTableClient returns a new table client
func NewTableClient(sqlClient *database.SQLClient) *TableClient {
	return &TableClient{sqlClient, database.Table{Name: tableName,
		Cols: Row{}.columnar().Columns(),
	}}
}

// CreateTable creates the table for storing DAG related information
func (client *TableClient) CreateTable() {
	client.sqlClient.CreateTable(client.tableDef)
}

func fmtSQLDate(dateStruct time.Time) string {
	return "'" + dateStruct.Format(dateutils.SQLiteDateForm) + "'"
}

// GetMetricsForDag retrieves the metrics rows for a given dag id
func (client *TableClient) GetMetricsForDag(dagName string, times ...time.Time) ([]Row, error) {
	result := newRowResult(0)
	timeLength := len(times)
	query := ""
	switch {
	case timeLength == 0:
		query = fmt.Sprintf(
			"SELECT * FROM metrics WHERE dag_name = '%s' ORDER BY %s ASC",
			dagName,
			metricsTimeName,
		)
	case timeLength == 1:
		startTime := times[0]
		query = fmt.Sprintf(
			"SELECT * FROM metrics WHERE dag_name = '%s' and %s > %s ORDER BY %s ASC",
			dagName,
			metricsTimeName,
			fmtSQLDate(startTime),
			metricsTimeName,
		)
	case timeLength == 2:
		startTime := times[0]
		endTime := times[1]
		query = fmt.Sprintf(
			"SELECT * FROM metrics WHERE dag_name = '%s' and %s between %s and %s ORDER BY %s ASC",
			dagName,
			metricsTimeName,
			fmtSQLDate(startTime),
			fmtSQLDate(endTime),
			metricsTimeName,
		)
	default:
		return nil, fmt.Errorf("Cannot have more that 2 times")
	}
	client.sqlClient.QueryIntoResults(
		&result,
		query,
	)
	return result.returnedRows, nil
}

// InsertMetric inserts the given metric row
func (client *TableClient) InsertMetric(metricRow Row) {
	client.sqlClient.Insert(tableName, metricRow.columnar())
}
