package database

import (
	"fmt"
	"goflow/internal/dateutils"
	"time"
)

// SQLType is a sql datatype with a name
type SQLType interface {
	typeName() string
	getValRep() string
}

// String is a string sql datatype
type String struct{ Val string }

func (s String) typeName() string {
	return "STRING"
}
func (s String) getValRep() string {
	return "'" + s.Val + "'"
}

// Int is an integer sql datatype
type Int struct{ Val int }

func (i Int) typeName() string {
	return "INT"
}
func (i Int) getValRep() string {
	return fmt.Sprint(i.Val)
}

// Int64 is a long in sql datatype
type Int64 struct{ Val int64 }

func (i Int64) typeName() string {
	return "BIGINT"
}
func (i Int64) getValRep() string {
	return fmt.Sprint(i.Val)
}

// TimeStamp is a time stamp sql datatype
type TimeStamp struct {
	Val time.Time
}

func (t TimeStamp) typeName() string {
	return "TIMESTAMP"
}
func (t TimeStamp) getValRep() string {
	return "'" + t.Val.Format(dateutils.SQLiteDateForm) + "'"
}

// Bool is a bool sql datatype
type Bool struct {
	Val bool
}

func (b Bool) typeName() string {
	return "BOOL"
}
func (b Bool) getValRep() string {
	return fmt.Sprint(b.Val)
}
