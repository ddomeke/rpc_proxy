package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnvFile loads environment variables from the specified .env file
func LoadEnvFile(envPath string) error {
	// Override the input path to use the fixed path
	envPath = "../.env"

	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return fmt.Errorf("could not resolve directory path: %v", err)
	}

	fmt.Printf("Loading .env file: %s\n", absPath)
	err = godotenv.Load(absPath)
	if err != nil {
		return fmt.Errorf("could not load .env file: %v", err)
	}

	fmt.Println(".env file successfully loaded")
	return nil
}

// InitLogger initializes the log file and returns it for later closing
func InitLogger() *os.File {
	logFile, err := os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
	}
	log.SetOutput(io.MultiWriter(logFile, os.Stdout)) // Write to console and file
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return logFile
}

// HexToUint64 is a helper function to convert hexadecimal string to number
func HexToUint64(hexStr string) (uint64, error) {
	// Remove "0x" prefix
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	// Convert hex string to number
	return strconv.ParseUint(hexStr, 16, 64)
}
