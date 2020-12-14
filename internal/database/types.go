package database

// SQLType is a sql datatype with a name
type SQLType interface {
	typeName() string
	getValRep(string) string
}

// String is a string sql datatype
type String struct{}

func (s String) typeName() string {
	return "STRING"
}

func (s String) getValRep(val string) string {
	return "'" + val + "'"
}

// Int is an integer sql datatype
type Int struct{}

func (i Int) typeName() string {
	return "INT"
}

func (i Int) getValRep(val string) string {
	return val
}

// TimeStamp is a time stamp sql datatype
type TimeStamp struct{}

func (t TimeStamp) typeName() string {
	return "TIMESTAMP"
}

func (t TimeStamp) getValRep(val string) string {
	return "'" + val + "'"
}
