package main

import (
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

func checkDiskUsage(cfg Config) {
	usage, err := disk.Usage(cfg.DiskPath)
	if err != nil {
		log.Printf("ERROR: Could not get disk usage for path %s: %v", cfg.DiskPath, err)
		return
	}

	percent := usage.UsedPercent
	totalGB := float64(usage.Total) / 1024 / 1024 / 1024
	freeGB := float64(usage.Free) / 1024 / 1024 / 1024
	isAlert := int(percent) >= cfg.Threshold

	// --- Logging Output ---
	log.Printf("--- Disk Usage Report for %s ---\n", cfg.DiskPath)
	log.Printf("Total: %.2f GB\n", totalGB)
	log.Printf("Free:  %.2f GB\n", freeGB)
	log.Printf("Used:  %.1f%%\n", percent)

	if isAlert {
		log.Printf("!!! ALERT: Disk usage (%.1f%%) exceeds threshold (%d%%) !!!\n", percent, cfg.Threshold)
	} else {
		log.Println("Disk usage is nominal.")
	}

	// --- Update Shared Status (Thread Safe) ---
	// Acquire write lock to ensure no web requests read incomplete data
	CurrentStatus.mu.Lock()

	CurrentStatus.TotalGB = totalGB
	CurrentStatus.FreeGB = freeGB
	CurrentStatus.UsedPercent = percent
	CurrentStatus.IsAlert = isAlert
	CurrentStatus.LastCheck = time.Now().Format("2006-01-02 15:04:05") // Standard Go time format

	// Release the lock
	CurrentStatus.mu.Unlock()
}
