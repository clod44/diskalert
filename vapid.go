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

type VAPIDKeys struct {
	PublicKey  string
	PrivateKey *ecdsa.PrivateKey
}

var GlobalVAPIDKeys VAPIDKeys

func encodeVAPIDKey(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}

func setupVAPIDKeys(cfg Config) {
	if err := os.MkdirAll(cfg.VapidDir, 0755); err != nil {
		log.Fatalf("Failed to create VAPID directory %s: %v", cfg.VapidDir, err)
	}

	privKeyPath := filepath.Join(cfg.VapidDir, cfg.VapidSecretKey)
	pubKeyPath := filepath.Join(cfg.VapidDir, cfg.VapidPublicKey)

	privateKeyData, err := os.ReadFile(privKeyPath)
	publicKeyData, pubErr := os.ReadFile(pubKeyPath)

	if err == nil && pubErr == nil {
		log.Printf("Found existing VAPID keys in %s. Loading them.", cfg.VapidDir)

		privateKey := new(ecdsa.PrivateKey)
		privateKey.Curve = elliptic.P256()

		dBytes, err := base64.URLEncoding.DecodeString(string(privateKeyData))
		if err != nil {
			log.Fatalf("Failed to decode VAPID private key from file: %v", err)
		}

		privateKey.D = new(big.Int).SetBytes(dBytes)
		
		privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.Curve.ScalarBaseMult(privateKey.D.Bytes())
		
		GlobalVAPIDKeys = VAPIDKeys{
			PublicKey:  string(publicKeyData),
			PrivateKey: privateKey,
		}
		return
	}

	log.Printf("VAPID keys not found. Generating new keys in %s.", cfg.VapidDir)

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate VAPID private key: %v", err)
	}

	privKeyBytes := privateKey.D.Bytes()
	privKeyEncoded := encodeVAPIDKey(privKeyBytes)

	pubKeyBytes := elliptic.Marshal(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyEncoded := encodeVAPIDKey(pubKeyBytes)

	if err := os.WriteFile(privKeyPath, []byte(privKeyEncoded), 0600); err != nil {
		log.Fatalf("Failed to save VAPID private key: %v", err)
	}
	if err := os.WriteFile(pubKeyPath, []byte(pubKeyEncoded), 0644); err != nil {
		log.Fatalf("Failed to save VAPID public key: %v", err)
	}

	GlobalVAPIDKeys = VAPIDKeys{
		PublicKey:  pubKeyEncoded,
		PrivateKey: privateKey,
	}

	log.Printf("Successfully generated and saved VAPID keys.")
}