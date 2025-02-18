package main

import (
	"bytes"
	"os"
	"path/filepath"
)

// FilterBinary returns true if the file appears to be a binary file.
func FilterBinary(node *FileNode) bool {
	// Skip directories
	if node.isDir {
		return false
	}

	// Check file extension first for common binary types
	ext := filepath.Ext(node.path)
	commonBinaryExts := map[string]bool{
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".o":     true,
		".a":     true,
		".class": true,
		".pyc":   true,
	}
	if commonBinaryExts[ext] {
		return true
	}

	// Check if file is executable on Unix systems
	if info, err := os.Stat(node.path); err == nil {
		if info.Mode()&0o111 != 0 { // Check if any execute bit is set
			return true
		}
	}

	// Read first 512 bytes to detect binary content
	file, err := os.Open(node.path)
	if err != nil {
		return false // If we can't read the file, assume it's not binary
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}
	buf = buf[:n]

	// Check for null bytes and high concentration of non-text bytes
	nullCount := bytes.Count(buf, []byte{0})
	if nullCount > 0 {
		return true
	}

	// Count non-printable characters (excluding common whitespace)
	nonPrintable := 0
	for _, b := range buf {
		if (b < 32 || b > 126) && !isWhitespace(b) {
			nonPrintable++
		}
	}

	// If more than 30% of content is non-printable, consider it binary
	return float64(nonPrintable)/float64(len(buf)) > 0.30
}

func isWhitespace(b byte) bool {
	return b == '\n' || b == '\r' || b == '\t' || b == ' '
}
