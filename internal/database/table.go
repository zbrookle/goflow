package database

import (
	"fmt"
)

const commaSpace = ", "

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
	Name          string
	Cols          []Column
	UniqueCols    []Column
	PrimaryKeyCol Column
}

// getCreateSyntax returns the create table segment of the create table expression
func (table *Table) getCreateSyntax() string {
	query := fmt.Sprintf("CREATE TABLE %s(", table.Name)
	colCount := len(table.Cols)
	for i, col := range table.Cols {
		query += col.String()
		if col == table.PrimaryKeyCol {
			query += " PRIMARY KEY"
		}
		if i < colCount-1 {
			query += commaSpace
		}
	}
	return query
}

func (table *Table) getUniqueSyntax() string {
	query := ""
	if table.UniqueCols != nil {
		query += ", UNIQUE("

		uniqueColCount := len(table.UniqueCols)
		for i, col := range table.UniqueCols {
			query += col.Name
			if i < uniqueColCount-1 {
				query += commaSpace
			}
		}
		query += ")"
	}
	return query + ")"
}

// createQuery returns the SQL query that can create the table represented by table
func (table *Table) createQuery() string {
	return table.getCreateSyntax() + table.getUniqueSyntax()
}
