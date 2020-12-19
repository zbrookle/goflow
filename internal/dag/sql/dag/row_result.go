package dag

import (
	"database/sql"
	"fmt"
	"goflow/internal/database"
	"time"
)

// Row is a struct containing data about a particular dag
type Row struct {
	ID              int
	Name            string
	Namespace       string
	Version         string
	FilePath        string
	FileFormat      string
	CreatedDate     time.Time
	LastUpdatedDate time.Time
}

// NewRow returns a new row with the appropriate update and create time stamps
func NewRow(id int, name, namespace, version, filePath, fileFormat string) Row {
	creation := time.Now()
	return Row{
		id, name, namespace, version, filePath, fileFormat, creation, creation,
	}
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
		row.ID,
		row.Name,
		row.Namespace,
		row.Version,
		row.FilePath,
		row.FileFormat,
		row.CreatedDate.String(),
		row.LastUpdatedDate.String(),
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
		{Column: database.Column{Name: "id", DType: database.Int{}}, Value: fmt.Sprint(row.ID)},
		{Column: database.Column{Name: nameName, DType: database.String{}}, Value: row.Name},
		{
			Column: database.Column{Name: namespaceName, DType: database.String{}},
			Value:  row.Namespace,
		},
		{Column: database.Column{Name: "version", DType: database.String{}}, Value: row.Version},
		{Column: database.Column{Name: "file_path", DType: database.String{}}, Value: row.FilePath},
		{
			Column: database.Column{Name: "file_format", DType: database.String{}},
			Value:  row.FileFormat,
		},
		{
			Column: database.Column{Name: "created_date", DType: database.TimeStamp{}},
			Value:  row.CreatedDate.String(),
		},
		{
			Column: database.Column{Name: "last_updated_date", DType: database.TimeStamp{}},
			Value:  row.LastUpdatedDate.String(),
		},
	}
}

func (result *dagRowResult) ScanAppend(rows *sql.Rows) error {
	row := Row{}
	err := rows.Scan(
		&row.ID,
		&row.Name,
		&row.Namespace,
		&row.Version,
		&row.FilePath,
		&row.FileFormat,
		&row.CreatedDate,
		&row.LastUpdatedDate,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}
