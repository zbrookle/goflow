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

// JSONPanicFormatBytes performs json.MarshalIndent
func JSONPanicFormatBytes(v interface{}) []byte {
	jsonPanic, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	return jsonPanic
}

// JSONPanicFormat performances json.MarshalIndent with built in panic and string conversion
func JSONPanicFormat(v interface{}) string {
	jsonString := JSONPanicFormatBytes(v)
	return string(jsonString)
}
