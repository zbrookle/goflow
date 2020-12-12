package database

import "fmt"

type column struct {
	name  string
	dtype string
}

func (col column) String() string {
	return fmt.Sprintf("%s, %s", col.name, col.dtype)
}

type table struct {
	name string
	cols []column
}

func (table *table) createQuery() string {
	query := fmt.Sprintf("CREATE TABLE %s(", table.name)
	for _, col := range table.cols {
		query += col.String()
	}
	return query
}
