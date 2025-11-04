package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var diskHistoryDB *sql.DB

// This is used as the input payload for the LogDiskRecord function.
type NewDiskRecord struct {
	DiskPath         string  // The mount point
	TotalSize        uint64 //Bytes
	UsedSize         uint64 //Bytes
	AvailableSize  	 uint64 
	UsedPercentage   float64 // 0-100
	UUID             string  
}

// This is used when READING data back from the database.
type DiskRecord struct {
	ID               int64
	Timestamp        int64   // Unix timestamp (seconds)
	DiskPath         string
	TotalSize        float64
	UsedSize         float64
	AvailableSize    float64
	UsedPercentage   float64
	UUID             string
	Forecast		 int64
}

func InitDiskHistoryDB() {
	if diskHistoryDB != nil {
		return
	}
	dbFileName := "history.db"
	dbPath := filepath.Join(getAppDir(), dbFileName)
	
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("FATAL: Failed to create database directory %s: %v", dbDir, err)
	}
	var err error
	diskHistoryDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("FATAL: Failed to open disk history database at %s: %v", dbPath, err)
	}

	if err := diskHistoryDB.Ping(); err != nil {
		log.Fatalf("FATAL: Failed to connect to disk history database: %v", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS DiskRecords (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp INTEGER NOT NULL,
        disk_path TEXT NOT NULL,
        uuid TEXT NOT NULL,
        total_size REAL NOT NULL,
        used_size REAL NOT NULL,
        available_size REAL NOT NULL,
        used_percentage REAL NOT NULL,
        forecast INTEGER DEFAULT 0, 
        UNIQUE(timestamp, uuid)
    );`
	if _, err := diskHistoryDB.Exec(createTableSQL); err != nil {
		log.Fatalf("FATAL: Failed to create DiskRecords table: %v", err)
	}
    createIndexSQL := `
    CREATE INDEX IF NOT EXISTS idx_uuid_time ON DiskRecords (uuid, timestamp DESC);`   
	if _, err := diskHistoryDB.Exec(createIndexSQL); err != nil {
		log.Fatalf("FATAL: Failed to create DiskRecords index: %v", err)
	}
	log.Printf("INFO: Disk History database initialized successfully at %s", dbPath)
}

func SaveDiskRecord(record NewDiskRecord) {
	if diskHistoryDB == nil {
		log.Println("ERROR: Cannot log disk record, database not initialized.")
		return
	}
	currentTimestamp := time.Now().Unix()

	query := `
	INSERT INTO DiskRecords (timestamp, disk_path, uuid, total_size, used_size, available_size, used_percentage) 
	VALUES (?, ?, ?, ?, ?, ?, ?);`

	_, err := diskHistoryDB.Exec(
		query,
		currentTimestamp,
		record.DiskPath,
		record.UUID,
		record.TotalSize,
		record.UsedSize,
		record.AvailableSize,
		record.UsedPercentage,
	)
	
	if err != nil {
		log.Printf("ERROR: Failed to log disk record for %s (%s): %v", record.DiskPath, record.UUID, err)
	}
}

// GetDiskRecords returns a list of disk records for the given UUID, sorted by timestamp in descending order.
// The limit parameter specifies the maximum number of records to return.
func GetDiskRecords(uuid string, limit int) ([]DiskRecord, error) {
	if diskHistoryDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
	SELECT id, timestamp, disk_path, uuid, total_size, used_size, available_size, used_percentage, forecast 
	FROM DiskRecords 
	WHERE uuid = ?
	ORDER BY timestamp DESC
	LIMIT ?`

	rows, err := diskHistoryDB.Query(query, uuid, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query disk records for UUID %s: %w", uuid, err)
	}
	defer rows.Close()

	var records []DiskRecord
	for rows.Next() {
		var r DiskRecord
		err := rows.Scan(&r.ID, &r.Timestamp, &r.DiskPath, &r.UUID, &r.TotalSize, &r.UsedSize, &r.AvailableSize, &r.UsedPercentage, &r.Forecast)
		if err != nil {
			log.Printf("WARN: Failed to scan DiskRecord row: %v", err)
			continue
		}
		records = append(records, r)
	}
    return records, rows.Err()
}
