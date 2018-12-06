package app

import (
	"log"
	"os"
)

var Logger *log.Logger

func SetupLogger(file *os.File) {
    Logger = log.New(file, "", log.Ltime|log.Llongfile)
}