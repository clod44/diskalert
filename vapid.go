package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"log"
	"math/big"
	"os"
	"path/filepath"
)

func encodeVAPIDKey(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}

func decodePrivateKey(encodedPrivKey string) (*ecdsa.PrivateKey, string) {
	dBytes, err := base64.URLEncoding.DecodeString(encodedPrivKey)
	if err != nil {
		log.Fatalf("Failed to decode VAPID private key: %v", err)
	}

	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = elliptic.P256()
	privateKey.D = new(big.Int).SetBytes(dBytes)
	privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.Curve.ScalarBaseMult(privateKey.D.Bytes())
	
	pubKeyBytes := elliptic.Marshal(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyEncoded := encodeVAPIDKey(pubKeyBytes)

	return privateKey, pubKeyEncoded
}

func setupVAPIDKeys() {
	privKeyPath, err := resolvePath(APP.cfg.VapidPrivate)
	if err != nil {
		log.Fatalf("Fatal path error for VapidPrivate: %v", err)
	}
	pubKeyPath, err := resolvePath(APP.cfg.VapidPublic)
	if err != nil {
		log.Fatalf("Fatal path error for VapidPublic: %v", err)
	}

	vapidDir := filepath.Dir(pubKeyPath)
	if err := os.MkdirAll(vapidDir, 0755); err != nil {
		log.Fatalf("Failed to create VAPID directory %s: %v", vapidDir, err)
	}

	var privKeyEncoded string
	writeToDisk := false

	if APP.vapidPrivateContent != "" {
		log.Printf("VAPID keys found in runtime variables. Using them and overwriting disk files.")
		privKeyEncoded = APP.vapidPrivateContent
		writeToDisk = true

	} else {
		privateKeyData, err := os.ReadFile(privKeyPath)
		if err == nil {
			log.Printf("Found existing VAPID private key on disk. Loading it.")
			privKeyEncoded = string(privateKeyData)
		} else {
			log.Printf("VAPID keys not found. Generating new keys in %s.", vapidDir)
			
			newPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				log.Fatalf("Failed to generate VAPID private key: %v", err)
			}
			
			privKeyEncoded = encodeVAPIDKey(newPrivateKey.D.Bytes())
			writeToDisk = true
		}
	}
	
	_, pubKeyEncoded := decodePrivateKey(privKeyEncoded)
	
	if writeToDisk {
		if err := os.WriteFile(privKeyPath, []byte(privKeyEncoded), 0600); err != nil {
			log.Fatalf("Failed to save VAPID private key: %v", err)
		}
		if err := os.WriteFile(pubKeyPath, []byte(pubKeyEncoded), 0644); err != nil {
			log.Fatalf("Failed to save VAPID public key: %v", err)
		}
		log.Printf("Successfully saved/updated VAPID keys to disk.")
	}

	APP.vapidPublic64 = pubKeyEncoded
	APP.vapidPrivateContent = privKeyEncoded
}