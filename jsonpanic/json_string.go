package jsonpanic

import "encoding/json"

// JSONPanic performances json.Marshal with built in panic
func JSONPanic(v interface{}) string {
	jsonString, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

// JSONPanicFormat performances json.MarshalIndent with built in panic
func JSONPanicFormat(v interface{}) string {
	jsonString, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}
