package dag

import (
	"database/sql"
	"fmt"
	"goflow/internal/database"
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

func (row Row) String() string {
	return fmt.Sprintf(
		`{
		  id: %d, 
		  name: %s, 
		  namespace: %s,
		  version: %s, 
		  filePath: %s,
		  fileFormat: %s, 
		  createDate: %s, 
		  lastUpdatedDate: %s
		}`,
		row.id,
		row.name,
		row.namespace,
		row.version,
		row.filePath,
		row.fileFormat,
		row.createdDate.String(),
		row.lastUpdatedDate.String(),
	)
}

type dagRowResult struct {
	returnedRows []Row
}

func newRowResult(n int) dagRowResult {
	return dagRowResult{
		returnedRows: make([]Row, 0, n),
	}
}

func (row Row) columnar() database.ColumnWithValueSlice {
	return []database.ColumnWithValue{
		{Column: database.Column{Name: "id", DType: database.Int{}}, Value: fmt.Sprint(row.id)},
		{Column: database.Column{Name: nameName, DType: database.String{}}, Value: row.name},
		{
			Column: database.Column{Name: namespaceName, DType: database.String{}},
			Value:  row.namespace,
		},
		{Column: database.Column{Name: "version", DType: database.String{}}, Value: row.version},
		{Column: database.Column{Name: "file_path", DType: database.String{}}, Value: row.filePath},
		{
			Column: database.Column{Name: "file_format", DType: database.String{}},
			Value:  row.fileFormat,
		},
		{
			Column: database.Column{Name: "created_date", DType: database.TimeStamp{}},
			Value:  row.createdDate.String(),
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

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}
