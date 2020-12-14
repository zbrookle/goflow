package dagrun

import (
	"database/sql"
	"time"
)

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
