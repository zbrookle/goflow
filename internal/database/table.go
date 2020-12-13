package database

import "fmt"

// Column is a column in the database structure
type Column struct {
	Name  string
	DType SQLType
}

func (col Column) String() string {
	return fmt.Sprintf("%s %s", col.Name, col.DType.typeName())
}

// Table can be used in various inputs to create tables
type Table struct {
	Name string
	Cols []Column
}

// createQuery returns the SQL query that can create the table represented by table
func (table *Table) createQuery() string {
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
