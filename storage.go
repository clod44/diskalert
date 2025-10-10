package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Keys struct {
	P256DH string `json:"p256dh"`
	Auth   string `json:"auth"`
}

type PushSubscription struct {
	Endpoint string `json:"endpoint"`
	Keys     Keys   `json:"keys"`
}

type SubscriptionStore struct {
	mu sync.RWMutex
	Subscriptions []PushSubscription
}

var SubscriptionDB = &SubscriptionStore{
	Subscriptions: make([]PushSubscription, 0),
}

func InitSubscriptionDB(cfg Config) {
	filePath := filepath.Join(cfg.VapidDir, cfg.SubscriptionFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("No existing subscription file found at %s. Initializing empty database.", filePath)
		return
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("ERROR: Failed to read subscription file %s: %v. Starting with empty database.", filePath, err)
		return
	}
	if len(data) == 0 {
		log.Println("Subscription file is empty. Initializing empty database.")
		return
	}
	var loadedSubs []PushSubscription
	if err := json.Unmarshal(data, &loadedSubs); err != nil {
		log.Printf("ERROR: Failed to unmarshal subscription data from %s: %v. Starting with empty database.", filePath, err)
		return
	}

	SubscriptionDB.mu.Lock()
	SubscriptionDB.Subscriptions = loadedSubs
	SubscriptionDB.mu.Unlock()

	log.Printf("Successfully loaded %d existing push subscriptions from %s.", len(loadedSubs), filePath)
}

func (s *SubscriptionStore) saveToFile(cfg Config) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.Subscriptions, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(cfg.VapidDir, cfg.SubscriptionFileName)
	return os.WriteFile(filePath, data, 0644)
}

func (s *SubscriptionStore) AddSubscription(sub PushSubscription, cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existingSub := range s.Subscriptions {
		if existingSub.Endpoint == sub.Endpoint {
			log.Printf("Subscription for endpoint %s already exists. Skipping.", sub.Endpoint)
			return nil
		}
	}

	s.Subscriptions = append(s.Subscriptions, sub)
	
	if err := s.saveToFile(cfg); err != nil {
		log.Printf("ERROR: Failed to save subscription file: %v", err)
	}
	log.Printf("New subscription added. Total active subscriptions: %d", len(s.Subscriptions))
	return nil
}