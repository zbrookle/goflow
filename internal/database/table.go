package database

import "fmt"

// Row is a row in a database, it must be defined differently depending on the table
// Values should returns a struct containing the values
// String should contain a string representation of the row, taking into account the schema, it should be of the form
// VALUES(val1, val2, val3,...)
type Row interface {
	Values()
	String() string
}

type column struct {
	name  string
	dtype string
}

func (col column) String() string {
	return fmt.Sprintf("%s %s", col.name, col.dtype)
}

type table struct {
	name string
	cols []column
}

func (table *table) createQuery() string {
	query := fmt.Sprintf("CREATE TABLE %s(", table.name)
	colCount := len(table.cols)
	for i, col := range table.cols {
		query += col.String()
		if i < colCount-1 {
			query += ", "
		}
	}
	query += ")"
	return query
}
