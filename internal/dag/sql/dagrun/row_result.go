package dagrun

import (
	"database/sql"
	"fmt"
	"goflow/internal/database"
	"time"
)

const statusName = "status"
const dagIDName = "dag_id"
const executionDateName = "execution_date"

// Row is a struct containing data about a particular dag
type Row struct {
	dagID           int
	status          string
	executionDate   time.Time
	startDate       time.Time
	endDate         time.Time
	lastUpdatedDate time.Time
}

type dagRowResult struct {
	rows         *sql.Rows
	returnedRows []Row
}

func newRowResult(n int) dagRowResult {
	return dagRowResult{
		returnedRows: make([]Row, 0, n),
	}
}

func (row Row) columnar() database.ColumnWithValueSlice {
	return []database.ColumnWithValue{
		{
			Column: database.Column{Name: dagIDName, DType: database.Int{}},
			Value:  fmt.Sprint(row.dagID),
		},
		{Column: database.Column{Name: statusName, DType: database.String{}}, Value: row.status},
		{
			Column: database.Column{Name: executionDateName, DType: database.TimeStamp{}},
			Value:  row.executionDate.String(),
		},
		{
			Column: database.Column{Name: "start_date", DType: database.TimeStamp{}},
			Value:  row.startDate.String(),
		},
		{
			Column: database.Column{Name: "end_date", DType: database.TimeStamp{}},
			Value:  row.endDate.String(),
		},
		{
			Column: database.Column{Name: "last_updated_date", DType: database.TimeStamp{}},
			Value:  row.lastUpdatedDate.String(),
		},
	}
}

func (result *dagRowResult) ScanAppend(rows *sql.Rows) error {
	row := Row{}
	err := rows.Scan(
		&row.dagID,
		&row.status,
		&row.executionDate,
		&row.startDate,
		&row.endDate,
		&row.lastUpdatedDate,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}
