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

func (col Column) WithValue(value string) ColumnWithValue {
	return ColumnWithValue{col, value}
}

// ColumnWithValue is a column with an associated value
type ColumnWithValue struct {
	Column
	Value string
}

// ValRep returns the values in it SQL representation
func (colWithVal ColumnWithValue) ValRep() string {
	return colWithVal.DType.getValRep(colWithVal.Value)
}

// getEqualsValue returns a string representing the conditional representation between the columns
func (colWithVal ColumnWithValue) getEqualsValue() string {
	return fmt.Sprintf("%s = %s", colWithVal.Name, colWithVal.ValRep())
}

// ColumnWithValueSlice is a slice of ColumnWithValue
type ColumnWithValueSlice []ColumnWithValue

// Join returns the equals value representation along separated by a given string
func (slice ColumnWithValueSlice) Join(sep string) string {
	result := ""
	for i, val := range slice {
		result += val.getEqualsValue()
		if i < len(slice)-1 {
			result += sep
		}
	}
	return result
}

// Columns returns a slice columns using the columns from the ColumnWithValue structs
func (slice ColumnWithValueSlice) Columns() []Column {
	columns := make([]Column, 0, len(slice))
	for _, colWithValue := range slice {
		columns = append(columns, colWithValue.Column)
	}
	return columns
}
