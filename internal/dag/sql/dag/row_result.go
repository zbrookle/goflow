package dag

import (
	"database/sql"
	"time"
)

// Row is a struct containing data about a particular dag
type Row struct {
	id              int
	name            string
	namespace       string
	version         string
	filePath        string
	fileFormat      string
	createdDate     time.Time
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

func (result *dagRowResult) ScanAppend() error {
	row := Row{}
	err := result.rows.Scan(
		&row.id,
		&row.name,
		&row.namespace,
		&row.version,
		&row.filePath,
		&row.fileFormat,
		&row.createdDate,
		&row.lastUpdatedDate,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *dagRowResult) Rows() *sql.Rows {
	return result.rows
}

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}

func (result *dagRowResult) SetRows(rows *sql.Rows) {
	result.rows = rows
}
