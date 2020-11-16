package podutils

import (
	"path/filepath"
	"runtime"
)

func getRootPath() string {
	_, filename, _, _ := runtime.Caller(0)
	fileNameAbs, err := filepath.Abs(filename)

	if err != nil {
		panic(err)
	}
	modPath := filepath.Dir(fileNameAbs)
	rootPath := filepath.Dir(modPath)
	return rootPath
}

// GetTestFolder returns the folder that the current test is running in
func GetTestFolder() string {
	return filepath.Join(getRootPath(), "test_files")
}

// GetConfigPath returns the file where the testing config is stored
func GetConfigPath() string {
	return filepath.Join(GetTestFolder(), "config.json")
}

// GetDagsFolder returns the folder where the test dags are stored
func GetDagsFolder() string {
	return filepath.Join(GetTestFolder(), "test_dags")
}
