package dags

import (
	"testing"

	"encoding/json"
)

type StringMap map[string]string

func map1InMap2(map1 StringMap, map2 StringMap) bool {
	for str := range map1 {
		if map1[str] != map2[str] {
			return false
		}
	}
	return true
}

func (stringMap StringMap) Equals(otherMap StringMap) bool {
	return map1InMap2(stringMap, otherMap) && map1InMap2(otherMap, stringMap) 
}

func (stringMap StringMap) Bytes() []byte {
	bytes, err := json.Marshal(stringMap)
	if err != nil {
		panic(err)
	}
	return bytes
}

func getMapFromJSONBytes(mapBytes []byte) StringMap {
	returnMap := make(StringMap)
	err := json.Unmarshal(mapBytes, &returnMap)
	if err != nil {
		panic(err)
	}
	return returnMap
}

func TestDAGFromJSONBytes(t *testing.T) {
	jsonMap := StringMap{"Name": "test", "Namespace": "default"}
	dagConfigBytes := jsonMap.Bytes()
	dag := createDAGFromJSONBytes(dagConfigBytes)
	marshaledJSON, err := json.Marshal(dag)
	if err != nil {
		panic(err)
	}
	if ! jsonMap.Equals(getMapFromJSONBytes(marshaledJSON)) {
		t.Error("DAG struct does not match up with expected values")
	}
}