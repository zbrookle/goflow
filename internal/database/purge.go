package database

import (
	"database/sql"
	"fmt"
	"goflow/internal/stringutils"
	"time"
)

type depQueryResult struct {
	rows                 *sql.Rows
	returnedRows         []foreignKeyTableRow
	hasUnlimitedCapacity bool
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

func (result *depQueryResult) HasUnlimitedCapacity() bool {
	return result.hasUnlimitedCapacity
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
					WHERE m.type = 'table' and p."table" = '%s'`,
			table),
	)
	tables := make([]string, 0, len(result.returnedRows))
	for _, row := range result.returnedRows {
		tables = append(tables, row.name)
	}
	return tables
}

// VerifyTableDrop returns true if the given table has been successfully dropped
func VerifyTableDrop(tableName string, client *SQLClient) bool {
	rows, err := client.Query(
		fmt.Sprintf("SELECT tbl_name FROM sqlite_master WHERE tbl_name = '%s'", tableName),
	)
	if err != nil {
		panic(err)
	}
	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	return rowCount != 0
}

// PurgeDB removes all tables from the database
func PurgeDB(client *SQLClient) {
	const dagRun = "dagrun"
	tables := client.Tables()
	fmt.Println("Tables Present", tables)
	tableSet := stringutils.NewStringSet(tables)
	if tableSet.Contains(dagRun) {
		fmt.Println("Dropping dagrun first")
		client.Exec(fmt.Sprintf("DROP TABLE %s", dagRun))
		tableSet.Remove(dagRun)
	}

	stack := make([]string, 0, len(tables))
	for len(tableSet) != 0 {
		table, err := tableSet.GetOne()
		if err != nil {
			panic(err)
		}
		stack = append(stack, table)
		for len(stack) > 0 {
			n := len(stack) - 1
			currTable := stack[n]
			stack = stack[:n]
			if tableSet.Contains(currTable) {
				dependents := getDependentTables(currTable, client)
				fmt.Println("Dependents of ", currTable, " are ", dependents)
				switch len(dependents) {
				case 0:
					fmt.Println("DROPPING", currTable)
					_, err := client.database.Exec(fmt.Sprintf("DROP TABLE %s", currTable))
					if err != nil {
						panic(err)
					}
					tableSet.Remove(currTable)
					for VerifyTableDrop(table, client) {
						time.Sleep(1 * time.Second)
					}
				default:
					stack = append(stack, currTable)
					stack = append(stack, dependents...)
				}
			}
		}
	}
}
