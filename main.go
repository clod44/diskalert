package main

import (
	"diskalert/diskmon"
	"diskalert/http"
	"diskalert/pkg/config"
	"diskalert/pkg/logger"
	"diskalert/pkg/service"
	"diskalert/storage/diskdb"
	"diskalert/storage/notifydb"
	"log"
	"time"
)



func main() {

	loggerFileHandle := logger.SetupLogger()
	if loggerFileHandle != nil {
		defer loggerFileHandle.Close()
	}


	config.SetupTLSFiles()     // this gotta be a blocking process so web server doesn't start before this
	notifydb.InitSubscriptionDB()
	diskdb.InitDiskHistoryDB()
	config.SetupVAPIDKeys()
	go http.StartWebServer() //"go" makes it a background process type shi without blocking the flow
	service.ManageServiceFile()
	
	interval := time.Duration(config.Cfg.CheckIntervalSeconds) * time.Second
	log.Printf("Monitoring %s every %d seconds. Threshold for all disks is %d%%.", config.Cfg.CheckIntervalSeconds, config.Cfg.Threshold)
	log.Println("--------------------------------------------------------------------------------")

	for {
		log.Println("Checking disk stats...")
		diskmon.UpdateDiskStatus() 
		logMonitoredDisks(diskmon.DiskStatus.Records)

		nextCheckTime := time.Now().Add(interval)
		log.Printf("Next check will be %d seconds later, at %s.", config.Cfg.CheckIntervalSeconds, nextCheckTime.Format("15:04:05"))
		time.Sleep(interval)
	}
}

func logMonitoredDisk(disk diskmon.MonitoredDisk) {
	log.Printf("  Disk: [%.1f%%] %s (%d bytes used / %d bytes total) (%s)", disk.UsedPercent, disk.DiskPath, disk.UsedSize, disk.TotalSize, disk.UUID)
}

func logMonitoredDisks(disks []diskmon.MonitoredDisk) (bool) {
	diskmon.DiskStatus.Mu.RLock()
	defer diskmon.DiskStatus.Mu.RUnlock()

	overallAlert := false
	
	log.Printf("Disk Usage Alert Report (Threshold: %d%%)", config.Cfg.Threshold)
	log.Println("--------------------------------------------------")

	for _, disk := range disks{
		if disk.IsAlert {
			overallAlert = true
			logMonitoredDisk(disk)
			log.Println("")
		}
	}

	if overallAlert {
		log.Println("--------------------------------------------------")
		log.Println("CRITICAL ALERT: One or more disks are over the usage threshold.")
	} else {
		log.Println("INFO: No disks currently in alert state.")
	}

	return overallAlert
}