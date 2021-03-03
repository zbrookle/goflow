package metrics

import (
	"database/sql"
	"goflow/internal/database"
	"goflow/internal/dateutils"
	"goflow/internal/jsonpanic"
	"time"
)

// Row is a struct containing data about a particular dag
type Row struct {
	ID               int
	DagName, PodName string
	Memory, CPU      int64
	MetricTime       time.Time
	CreatedDate      time.Time
	LastUpdatedDate  time.Time
}

// IDName is the column name for the primary id column
const IDName = "id"
const metricsTimeName = "metrics_time"

// NewRow returns a new row with the appropriate update and create time stamps
func NewRow(id int, dagName, podName string, memory, cpu int64, metricTime time.Time) Row {
	creationTime := dateutils.GetDateTimeNowMilliSecond()
	return Row{
		id, dagName, podName, memory, cpu, metricTime, creationTime, creationTime,
	}
}

func (row Row) String() string {
	return jsonpanic.JSONPanicFormat(row)
}

type dagRowResult struct {
	returnedRows         []Row
	hasUnlimitedCapacity bool
}

func newRowResult(n int) dagRowResult {

	result := dagRowResult{
		returnedRows: make([]Row, 0, n), hasUnlimitedCapacity: n == 0,
	}
	return result
}

func (row Row) columnar() database.ColumnWithValueSlice {
	return []database.ColumnWithValue{
		{Column: database.Column{Name: IDName, DType: database.Int{Val: row.ID}}},
		{Column: database.Column{Name: "dag_name", DType: database.String{Val: row.DagName}}},
		{Column: database.Column{Name: "pod_name", DType: database.String{Val: row.PodName}}},
		{Column: database.Column{Name: "memory", DType: database.Int64{Val: row.Memory}}},
		{Column: database.Column{Name: "cpu", DType: database.Int64{Val: row.CPU}}},
		{
			Column: database.Column{
				Name:  metricsTimeName,
				DType: database.TimeStamp{Val: row.MetricTime},
			},
		},
		{
			Column: database.Column{
				Name:  "created_date",
				DType: database.TimeStamp{Val: row.CreatedDate},
			},
		},
		{
			Column: database.Column{
				Name:  "last_updated_date",
				DType: database.TimeStamp{Val: row.LastUpdatedDate},
			},
		},
	}
}

func (result *dagRowResult) ScanAppend(rows *sql.Rows) error {
	row := Row{}
	err := rows.Scan(
		&row.ID,
		&row.DagName,
		&row.PodName,
		&row.Memory,
		&row.CPU,
		&row.MetricTime,
		&row.CreatedDate,
		&row.LastUpdatedDate,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}
func (result *dagRowResult) HasUnlimitedCapacity() bool {
	return result.hasUnlimitedCapacity
}
