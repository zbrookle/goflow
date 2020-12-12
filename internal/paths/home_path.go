package paths

import (
	"os"
	"path"
)

// GetGoDefaultHomePath returns the default location for the goflow config file
func GetGoDefaultHomePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(homeDir, ".goflow", "config.json")
}
