package logs

import (
	"log"
	"os"
)

var (
	// WarningLogger is the logger for warnings
	WarningLogger *log.Logger

	// InfoLogger is the logger for info
	InfoLogger *log.Logger

	// ErrorLogger is the logger for errors
	ErrorLogger *log.Logger
)

func init() {
	_, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
