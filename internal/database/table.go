package database

import "fmt"

// Column is a column in the database structure
type Column struct {
	name  string
	dtype string
}

func (col Column) String() string {
	return fmt.Sprintf("%s %s", col.name, col.dtype)
}

// Table can be used in various inputs to create tables
type Table struct {
	Name string
	Cols []Column
}

// CreateQuery returns the SQL query that can create the table represented by table
func (table *Table) CreateQuery() string {
	query := fmt.Sprintf("CREATE TABLE %s(", table.Name)
	colCount := len(table.Cols)
	for i, col := range table.Cols {
		query += col.String()
		if i < colCount-1 {
			query += ", "
		}
	}
	query += ")"
	return query
}
