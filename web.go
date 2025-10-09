package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	CurrentStatus.mu.RLock()
	defer CurrentStatus.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(CurrentStatus); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to encode metrics to JSON: %v", err)
	}
}

func startWebServer(cfg Config) {
	var certFilePath = filepath.Join(cfg.CertDir, cfg.CertFileName)
	var keyFilePath = filepath.Join(cfg.CertDir, cfg.KeyFileName)

	http.HandleFunc("/api/metrics", metricsHandler)

	publicFS, err := fs.Sub(Assets, "public")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem for 'public' (check if public folder is correctly embedded): %v", err)
	}
	fileServer := http.FileServer(http.FS(publicFS))
	http.Handle("/", fileServer)

	bindAddress := fmt.Sprintf(":%d", cfg.Port)

	log.Println("========================================================================================")
	log.Printf("DiskAlert Web Dashboard is now LIVE!")
	log.Printf("Access Dashboard at: http://<YOUR_IP>:%d/", cfg.Port)
	log.Printf("JSON Metrics API:    http://<YOUR_IP>:%d/api/metrics", cfg.Port)
	log.Println("========================================================================================")

	err2 := http.ListenAndServeTLS(bindAddress, certFilePath, keyFilePath, nil)
	if err != nil {
		log.Fatalf("HTTPS Web server failed to start on port %d: %v", cfg.Port, err2)
	}
}
