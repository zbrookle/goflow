package database

// SQLType is a sql datatype with a name
type SQLType interface {
	typeName() string
}

// String is a string sql datatype
type String struct{}

func (s String) typeName() string {
	return "STRING"
}

// Int is an integer sql datatype
type Int struct{}

func (i Int) typeName() string {
	return "INT"
}

// TimeStamp is a time stamp sql datatype
type TimeStamp struct{}

func (t TimeStamp) typeName() string {
	return "TIMESTAMP"
}
