package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func setupLogger(cfg Config) *os.File {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	if !cfg.LogToFile {
		log.SetOutput(os.Stderr)
		log.Println("Logger initialized. Logging to console only (LogToFile disabled in config).")
		return nil
	}

	appDir := getAppDir()
	finalLogPath := ""

	if filepath.IsAbs(cfg.LogFilePath) {
		finalLogPath = filepath.Join(cfg.LogFilePath, logFileName)
	} else if cfg.LogFilePath == "./" {
		finalLogPath = filepath.Join(appDir, logFileName)
	} else {
		finalLogPath = filepath.Join(appDir, cfg.LogFilePath, logFileName)
	}

	logFile, err := os.OpenFile(finalLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file at %s: %v", finalLogPath, err)
	}

	multiWriter := io.MultiWriter(os.Stderr, logFile)

	log.SetOutput(multiWriter)
	log.Println("Utility Started. Logging output redirected to file and console.")

	return logFile
}
