//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")
	os.Setenv("CGO_ENABLED", "0")

	outputFile := "diskalert"

	fmt.Println("--- Starting Go Cross-Compilation ---")
	fmt.Printf("Target OS/Arch: %s/%s\n", os.Getenv("GOOS"), os.Getenv("GOARCH"))
	fmt.Printf("Output File: %s\n", outputFile)
	fmt.Println("-------------------------------------")

	fmt.Println("Running 'go mod tidy' to download required dependencies...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		log.Fatalf("ERROR: 'go mod tidy' failed: %v", err)
	}
	fmt.Println("Dependencies resolved successfully.")

	fmt.Println("\nRunning 'go build'...")
	buildCmd := exec.Command("go", "build", "-v", "-o", outputFile, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		log.Fatalf("ERROR: Build failed: %v", err)
	}

	fmt.Println("\n--- Build Complete ---")
	fmt.Printf("Success! Binary created: %s\n", outputFile)
}
