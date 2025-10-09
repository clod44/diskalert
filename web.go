package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	http.HandleFunc("/api/metrics", metricsHandler)

	fileServer := http.FileServer(http.FS(Assets))
	// to always use the contents of "public" at / rather than public folder itself
	http.Handle("/", http.StripPrefix("/", fileServer))

	bindAddress := fmt.Sprintf(":%d", cfg.Port)

	log.Println("========================================================================================")
	log.Printf("DiskAlert Web Dashboard is now LIVE!")
	log.Printf("Access Dashboard at: http://<PACS_IP>:%d/", cfg.Port)
	log.Printf("JSON Metrics API:    http://<PACS_IP>:%d/api/metrics", cfg.Port)
	log.Println("========================================================================================")

	if err := http.ListenAndServe(bindAddress, nil); err != nil {
		log.Fatalf("Web server failed to start on port %d: %v", cfg.Port, err)
	}
}
