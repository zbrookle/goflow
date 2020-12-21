package dagrun

import (
	"database/sql"
	"goflow/internal/database"
	"goflow/internal/dateutils"
	"goflow/internal/jsonpanic"
	"time"
)

const statusName = "status"
const dagIDName = "dag_id"
const executionDateName = "execution_date"

// Row is a struct containing data about a particular dag
type Row struct {
	DagID           int
	Status          string
	ExecutionDate   time.Time
	StartDate       time.Time
	EndDate         time.Time
	LastUpdatedDate time.Time
}

func (row Row) String() string {
	return jsonpanic.JSONPanicFormat(row)
}

// NewRow returns a new row with the appropriate update and create time stamps
func NewRow(dagID int, status string, executionDate time.Time) Row {
	creationTime := dateutils.GetDateTimeNowMilliSecond()
	return Row{
		DagID: dagID, Status: status, ExecutionDate: executionDate, StartDate: creationTime, LastUpdatedDate: creationTime,
	}
}

// RowList is a list of Rows
type RowList []Row

// Less returns true if the ExecutionDate of the row at i comes before the one at j
func (rl RowList) Less(i, j int) bool {
	return rl[i].ExecutionDate.Before(rl[j].ExecutionDate)
}

// Len returns the length of the row
func (rl RowList) Len() int {
	return len(rl)
}

// Swap switches the indices of the two elements
func (rl RowList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

type dagRowResult struct {
	rows         *sql.Rows
	returnedRows RowList
}

func newRowResult(n int) dagRowResult {
	return dagRowResult{
		returnedRows: make([]Row, 0, n),
	}
}

func (row Row) columnar() database.ColumnWithValueSlice {
	return []database.ColumnWithValue{
		{
			Column: database.Column{Name: dagIDName, DType: database.Int{Val: row.DagID}},
		},
		{Column: database.Column{Name: statusName, DType: database.String{Val: row.Status}}},
		{
			Column: database.Column{
				Name:  executionDateName,
				DType: database.TimeStamp{Val: row.ExecutionDate},
			},
		},
		{
			Column: database.Column{
				Name:  "start_date",
				DType: database.TimeStamp{Val: row.StartDate},
			},
		},
		{
			Column: database.Column{Name: "end_date", DType: database.TimeStamp{Val: row.EndDate}},
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
		&row.DagID,
		&row.Status,
		&row.ExecutionDate,
		&row.StartDate,
		&row.EndDate,
		&row.LastUpdatedDate,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *dagRowResult) Capacity() int {
	return cap(result.returnedRows)
}
