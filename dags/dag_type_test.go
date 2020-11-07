package dags

import (
	"path/filepath"
	"sort"
	"testing"

	"encoding/json"
	"runtime"
)

var DAGPATH string

func getTestFolder() string {
    _, filename, _, _ := runtime.Caller(0)
	fileNameAbs, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}
	return filepath.Dir(fileNameAbs)
}

func TestMain(m *testing.M) {
	DAGPATH = filepath.Join(getTestFolder(), "test_dags")
	m.Run()
}

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

func TestReadFiles(t *testing.T){ 
	expectedFiles := []string{"my_json_dag.json", "my_json_dag2.json", "my_python_dag.py"}
	sort.Strings(expectedFiles)
	foundFilePaths := getDirSliceRecur(DAGPATH)
	for i, filePath := range foundFilePaths {
		_, foundFilePaths[i] = filepath.Split(filePath)
	}
	sort.Strings(foundFilePaths)
	expectedFileCount := len(expectedFiles)
	foundFileCount := len(foundFilePaths)
	if len(expectedFiles) != len(foundFilePaths) {
		t.Errorf("Expected %d files, found %d files", expectedFileCount, foundFileCount)
		panic("File counts are different")
	}
	for i, foundPath := range foundFilePaths {
		expectedFile := expectedFiles[i] 
		_, foundFile := filepath.Split(foundPath)
		if expectedFiles[i] != foundFile {
			t.Errorf("Expected file %s, found file %s", expectedFile, foundFile)
		}
	}
}
