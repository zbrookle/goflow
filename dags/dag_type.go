package dags

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DAG is directed Acyclic graph for hold job information
type DAG struct {
	Name        string
	Namespace   string
	Schedule    string
	DockerImage string
	RetryPolicy string
}

func readDAGFile(dagFilePath string) []byte {
	dat, err := ioutil.ReadFile(dagFilePath)
	if err != nil {
		panic(err)
	}
	return dat
}

func createDAGFromJSONBytes(dagBytes []byte) DAG {
	dagStruct := DAG{}
	err := json.Unmarshal(dagBytes, &dagStruct)
	if err != nil {
		panic(err)
	}
	return dagStruct
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(dagFilePath string) DAG {
	dagBytes := readDAGFile(dagFilePath)
	return createDAGFromJSONBytes(dagBytes)
}

// getDirSliceRecur recursively retrieves all file names from the directory
func getDirSliceRecur(directory string) []string {
	files := []string{}
	dagFileRegex := regexp.MustCompile(".*_dag.*\\.(go|json|py)")
	appendToFiles := func(path string, info os.FileInfo, err error) error { 
		if err != nil {
			return err
		}
		if dagFileRegex.Match([]byte(path)) {
			files = append(files, path)
		}
		return nil
	}
	err := filepath.Walk(directory, appendToFiles)
	if err != nil {
		panic(err)
	}
	return files
}

// GetDAGSFromFolder returns a slice of DAG structs, one for each DAG file
// Each file must have the "dag" suffix
// E.g., my_dag.py, some_dag.json
func GetDAGSFromFolder(folder string) []DAG {
	files := getDirSliceRecur(folder)
	dags := make([]DAG, 0, len(files))
	for _, file := range(files) {
		if strings.ToLower(filepath.Ext(file)) == "json" {
			dags = append(dags, getDAGFromJSON(file))
		}
	}
	return dags
}