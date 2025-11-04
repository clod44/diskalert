package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type DiskStatus struct {
	mu       sync.RWMutex
	Records   []MonitoredDisk `json:"records"`
	LastCheck string       `json:"last_check"`
	IsAlert    bool          `json:"is_alert"`
}

type MonitoredDisk struct {
	DiskPath   string `json:"disk_path"`
	UUID      string `json:"uuid"`
	TotalSize uint64 `json:"total_size"`
	UsedSize   uint64 `json:"used_size"`
	UsedPercent float64 `json:"used_percent"`
	IsAlert    bool   `json:"is_alert"`
}

var DiskUUIDMap = make(map[string]string)

func findUUIDForDevice(devicePath string) string {
	if identifier, ok := DiskUUIDMap[devicePath]; ok {
		return identifier
	}

	uuid, err := getUUIDFromSymlinks(devicePath)
	if err == nil && uuid != "" {
		DiskUUIDMap[devicePath] = uuid
		return uuid
	}

	DiskUUIDMap[devicePath] = devicePath
	return devicePath
}

func getUUIDFromSymlinks(devicePath string) (string, error) {
	deviceBase := filepath.Base(devicePath)
	
	dirEntries, err := os.ReadDir("/dev/disk/by-uuid")
	if err != nil {
		return "", err
	}

	for _, entry := range dirEntries {
		if entry.Type()&os.ModeSymlink != 0 {
			symlinkPath := filepath.Join("/dev/disk/by-uuid", entry.Name())
			targetPath, err := os.Readlink(symlinkPath)
			if err != nil {
				continue
			}

			cleanTarget := filepath.Clean(targetPath)
			targetBase := filepath.Base(cleanTarget)

			if targetBase == deviceBase {
				return entry.Name(), nil
			}
		}
	}

	return "", fmt.Errorf("UUID not found for device %s", devicePath)
}

func isExcluded(path string) bool {
	for _, excludedPath := range APP.cfg.ExcludePaths {
		if path == excludedPath {
			return true
		}
	}
	return false
}

func UpdateDiskStatus() {
	partitions, err := disk.Partitions(true)
	if err != nil {
		log.Printf("ERROR: Failed to list disk partitions: %v", err)
		return
	}

	var currentMonitors []MonitoredDisk

	for _, p := range partitions {
		mountPoint := p.Mountpoint

		if isExcluded(mountPoint) {
			continue
		}

		usage, err := disk.Usage(mountPoint)
		if err != nil {
			log.Printf("ERROR: Could not get disk usage for mount point %s: %v", mountPoint, err)
			continue
		}
		
		if usage.Total == 0 || usage.Fstype == "tmpfs" || usage.Fstype == "devtmpfs" {
			continue
		}
		
		percent := usage.UsedPercent
		totalBytes := usage.Total
		usedBytes := usage.Used
		isAlert := int(percent) >= APP.cfg.Threshold
		
		diskUUID := findUUIDForDevice(p.Device)
		
		record := NewDiskRecord{
			DiskPath:      mountPoint,
			UUID:         diskUUID,
			TotalSize:   totalBytes,
			UsedSize:   usedBytes,
			AvailableSize: totalBytes - usedBytes,
			UsedPercentage: percent,
		}

		SaveDiskRecord(record)

		log.Printf("--- Disk Report: %s (UUID: %s) ---", mountPoint, diskUUID)
		log.Printf("Total: %d bytes", totalBytes)
		log.Printf("Used:   %d bytes (%.1f%%)", usedBytes, percent)
		if isAlert {
			log.Printf("!!! ALERT: Usage exceeds threshold (%d%%) !!!", APP.cfg.Threshold)
		}

		currentMonitors = append(currentMonitors, MonitoredDisk{
			DiskPath:   mountPoint,
			UUID:      diskUUID,
			TotalSize: totalBytes,
			UsedSize:   usedBytes,
			UsedPercent: percent,
			IsAlert:    isAlert,
		})
	}

	APP.diskStatus.mu.Lock()
	APP.diskStatus.Records = currentMonitors
	APP.diskStatus.LastCheck = time.Now().Format("2006-01-02 15:04:05")
	APP.diskStatus.mu.Unlock()
}