package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
)

func setupLogger() *os.File {
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	if !APP.cfg.LogToFile {
		log.SetOutput(os.Stderr)
		log.Println("Logger initialized. Logging to console only (LogToFile disabled in config).")
		return nil
	}

	finalLogPath := resolveConfigPath(APP.cfg.LogFile)

	logDir := filepath.Dir(finalLogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("WARNING: Failed to create log directory %s: %v. Logging only to console.", logDir, err)
		log.SetOutput(os.Stderr)
		return nil
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

func logJson(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("ERROR: Failed to format data for JSON logging: %v", err)
		return
	}
	log.Println(string(jsonBytes))
}