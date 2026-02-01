package utils

import (
	"log"
	"os"
)

// TrimString truncates a string to a maximum length.
func TrimString(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// Ckerr checks for an error and bails on failure.
func Ckerr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// CloseFileBOF closes the file and bails on failure.
func CloseFileBOF(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Fatalf("Error closing file: %v", err)
	}
}
