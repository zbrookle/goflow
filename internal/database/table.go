package database

import (
	"fmt"
)

const commaSpace = ", "

// KeyReference is a pair of columns defining that one column references another
type KeyReference struct {
	Key      Column
	RefTable string
	RefCol   Column
}

// refString returns a string of SQL that reflects the KeyReference struct
func (ref KeyReference) refString() string {
	return fmt.Sprintf(
		"FOREIGN KEY(%s) REFERENCES %s(%s)",
		ref.Key.Name,
		ref.RefTable,
		ref.RefCol.Name,
	)
}

// Table can be used in various inputs to create tables
type Table struct {
	Name          string
	Cols          []Column
	UniqueCols    []Column
	PrimaryKeyCol Column
	ForeignKeys   []KeyReference
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
	return query
}

func (table *Table) getReferenceSyntax() string {
	query := ""
	if table.ForeignKeys != nil {
		for _, col := range table.ForeignKeys {
			query += ", " + col.refString()
		}
	}
	return query
}

// createQuery returns the SQL query that can create the table represented by table
func (table *Table) createQuery() string {
	return table.getCreateSyntax() + table.getUniqueSyntax() + table.getReferenceSyntax() + ")"
}

// GetColumnsWithValues returns a slice of ColumnWithValue struct with the given values
func (table *Table) GetColumnsWithValues(values []string) []ColumnWithValue {
	colLength := len(table.Cols)
	valLength := len(values)
	if colLength != valLength {
		panic(fmt.Sprintf("table can only recieve %d values, was given %d", colLength, valLength))
	}
	columnsWithValues := make([]ColumnWithValue, 0, len(values))
	for i, col := range table.Cols {
		columnsWithValues = append(columnsWithValues, col.WithValue(values[i]))
	}
	return columnsWithValues
}
