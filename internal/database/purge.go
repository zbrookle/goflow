package database

import (
	"database/sql"
	"fmt"
	"goflow/internal/stringutils"
)

type depQueryResult struct {
	rows         *sql.Rows
	returnedRows []foreignKeyTableRow
}

type foreignKeyTableRow struct {
	name string
}

func (result *depQueryResult) ScanAppend(rows *sql.Rows) error {
	row := foreignKeyTableRow{}
	err := rows.Scan(
		&row.name,
	)
	result.returnedRows = append(result.returnedRows, row)
	return err
}

func (result *depQueryResult) Capacity() int {
	return cap(result.returnedRows)
}

// getDependentTables returns a slice of table names that are dependent on this table
func getDependentTables(table string, client *SQLClient) []string {
	result := depQueryResult{}
	client.QueryIntoResults(
		&result,
		fmt.Sprintf(`SELECT 
						m.name
					FROM
						sqlite_master m
						JOIN pragma_foreign_key_list(m.name) p 
						ON m.name != p."table"
					WHERE m.type = 'table' and p."table" = '%s'
					ORDER BY m.name`,
			table),
	)
	tables := make([]string, 0, len(result.returnedRows))
	for _, row := range result.returnedRows {
		tables = append(tables, row.name)
	}
	return tables
}

// PurgeDB removes all tables from the database
func PurgeDB(client *SQLClient) {
	tables := client.Tables()
	tableSet := stringutils.NewStringSet(tables)

	stack := make([]string, 0, len(tables))
	for len(tableSet) != 0 {
		table, err := tableSet.GetOne()
		if err != nil {
			panic(err)
		}
		stack = append(stack, table)
		fmt.Println(stack)
		for len(stack) > 0 {
			n := len(stack) - 1
			currTable := stack[n]
			stack = stack[:n]
			if _, ok := tableSet[currTable]; ok {
				dependents := getDependentTables(currTable, client)
				switch len(dependents) {
				case 0:
					_, err := client.database.Exec(fmt.Sprintf("DROP TABLE %s", table))
					if err != nil {
						panic(err)
					}
					tableSet.Remove(currTable)
					fmt.Println(tableSet)
				default:
					stack = append(stack, currTable)
					stack = append(stack, dependents...)
				}
			}
		}
	}
}
