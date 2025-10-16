package main

import (
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type DiskStatus struct {
	mu          sync.RWMutex
	DiskPath    string  `json:"disk_path"`
	Threshold   int     `json:"threshold"`
	TotalGB     float64 `json:"total_gb"`
	FreeGB      float64 `json:"free_gb"`
	UsedPercent float64 `json:"used_percent"`
	IsAlert     bool    `json:"is_alert"`
	LastCheck   string  `json:"last_check"`
}

func UpdateDiskStatus() {
	usage, err := disk.Usage(APP.cfg.DiskPath)
	if err != nil {
		log.Printf("ERROR: Could not get disk usage for path %s: %v", APP.cfg.DiskPath, err)
		return
	}

	percent := usage.UsedPercent
	totalGB := float64(usage.Total) / 1024 / 1024 / 1024
	freeGB := float64(usage.Free) / 1024 / 1024 / 1024
	isAlert := int(percent) >= APP.cfg.Threshold

	// --- Logging Output ---
	log.Printf("--- Disk Usage Report for %s ---\n", APP.cfg.DiskPath)
	log.Printf("Total: %.2f GB\n", totalGB)
	log.Printf("Free:  %.2f GB\n", freeGB)
	log.Printf("Used:  %.1f%%\n", percent)

	if isAlert {
		log.Printf("!!! ALERT: Disk usage (%.1f%%) exceeds threshold (%d%%) !!!\n", percent, APP.cfg.Threshold)
	} else {
		log.Println("Disk usage is nominal.")
	}

	// --- Update Shared Status (Thread Safe) ---
	// Acquire write lock to ensure no web requests read incomplete data
	APP.diskStatus.mu.Lock()

	APP.diskStatus.TotalGB = totalGB
	APP.diskStatus.FreeGB = freeGB
	APP.diskStatus.UsedPercent = percent
	APP.diskStatus.IsAlert = isAlert
	APP.diskStatus.LastCheck = time.Now().Format("2006-01-02 15:04:05") // Standard Go time format

	// Release the lock
	APP.diskStatus.mu.Unlock()
}
