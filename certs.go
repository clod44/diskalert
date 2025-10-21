package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)



func validateExistingCert(certPath string, expectedIP string) bool {
	data, err := os.ReadFile(certPath)
	if err != nil {
		log.Printf("Error reading certificate file %s for validation: %v", certPath, err)
		return false
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		log.Printf("Error decoding PEM block from certificate file %s.", certPath)
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Printf("Error parsing certificate from file %s: %v", certPath, err)
		return false
	}

	if time.Now().After(cert.NotAfter) {
		log.Printf("CRITICAL: Certificate in %s expired on %s. Regeneration required.", certPath, cert.NotAfter.Format(time.RFC822))
		return false
	}

	targetIP := net.ParseIP(expectedIP)
	if targetIP == nil {
		log.Printf("CRITICAL: Configured IP address '%s' in the config file is not a valid IP address for certificate validation.", expectedIP)
		return false
	}

	for _, ip := range cert.IPAddresses {
		if ip.Equal(targetIP) {
			return true
		}
	}

	log.Printf("Validation failed: Certificate does not contain required IP %s in its SAN list.", expectedIP)
	return false
}

func SetupTLSFiles() {
	certFilePath, err := resolvePath(APP.cfg.CertFile)
	if err != nil {
		log.Fatalf("Fatal path error for CertFile: %v", err)
	}
	keyFilePath, err := resolvePath(APP.cfg.KeyFile)
	if err != nil {
		log.Fatalf("Fatal path error for KeyFile: %v", err)
	}

	certDir := filepath.Dir(certFilePath) 

	if err := os.MkdirAll(certDir, 0755); err != nil { 
		log.Fatalf("Failed to create certificate directory %s: %v", certDir, err)
	}

	if _, err := os.Stat(certFilePath); err == nil {
		if _, err := os.Stat(keyFilePath); err == nil {
			
			log.Printf("Found existing TLS files in %s. Using them.", certDir)
			
			isIPValid := validateExistingCert(certFilePath, APP.cfg.IP)
			
			if !isIPValid {
				log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				log.Printf("!! CRITICAL SECURITY ERROR: CERTIFICATE STALE !!")
				log.Printf("!! The IP address configured (%s) is NOT present in the existing certificate's SAN list, OR the certificate has expired.", APP.cfg.IP)
				log.Printf("!! WARNING: The web server will start, but connections from %s will FAIL with 'unknown certificate' errors.", APP.cfg.IP)
				log.Printf("!! ACTION REQUIRED: To fix this, stop the app, manually delete %s and %s, and restart.", certFilePath, keyFilePath)
				log.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			}
			return
		}
	}

	log.Println("TLS files not found. Generating new self-signed certificate...")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	ipList := []net.IP{net.ParseIP("127.0.0.1")}
    commonName := "diskalert.local"

	if APP.cfg.IP != "" {
		configuredIP := net.ParseIP(APP.cfg.IP)
		if configuredIP == nil {
			log.Fatalf("Configured IP address '%s' is invalid. Please check the config file.", APP.cfg.IP)
		}
		ipList = append(ipList, configuredIP)
	} else {
		log.Printf("WARNING: Configuration field 'IP' is empty. Certificate only valid for 127.0.0.1 and '%s'.", commonName)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"DiskAlert Self-Signed CA"},
			CommonName:   commonName,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           ipList,
        DNSNames: []string{commonName},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create(certFilePath)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", certFilePath, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to encode certificate: %v", err)
	}
	certOut.Close()

	keyOut, err := os.OpenFile(keyFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", keyFilePath, err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		log.Fatalf("Failed to encode private key: %v", err)
	}
	keyOut.Close()

	log.Printf("Successfully generated and saved self-signed TLS files to %s and %s.", certFilePath, keyFilePath)
}