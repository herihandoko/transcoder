package utils

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SanitizeFilename converts filename to safe directory name
func SanitizeFilename(filename string) string {
	// Remove file extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	sanitized := reg.ReplaceAllString(name, "-")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "video"
	}

	return sanitized
}

// GenerateTranscodedPath generates path for transcoded files with date structure
func GenerateTranscodedPath(basePath, filename string, resolution string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	dirName := SanitizeFilename(filename)

	return filepath.Join(basePath, year, month, dirName, resolution)
}

// GenerateArchivePath generates path for archived original files
func GenerateArchivePath(basePath, filename string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	return filepath.Join(basePath, year, month)
}
