package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linier-channel/internal/config"
	"linier-channel/internal/database"
	"linier-channel/internal/models"
	"linier-channel/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <video_id>", os.Args[0])
	}

	videoIDStr := os.Args[1]
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		log.Fatalf("Invalid video ID: %v", err)
	}

	cfg := config.Load()

	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get video and profiles
	var video models.Video
	if err := db.Preload("VideoProfiles").First(&video, uint(videoID)).Error; err != nil {
		log.Fatalf("Failed to get video: %v", err)
	}

	// Generate master playlist content
	masterPlaylist := generateMasterPlaylistContent(video.VideoProfiles)
	
	// Save master playlist to file
	dirName := utils.SanitizeFilename(video.OriginalFilename)
	// Use relative path for local development
	videoDir := filepath.Join("./transcoded-videos", dirName)
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	masterPath := filepath.Join(videoDir, "master.m3u8")
	if err := ioutil.WriteFile(masterPath, []byte(masterPlaylist), 0644); err != nil {
		log.Fatalf("Failed to write master playlist: %v", err)
	}

	fmt.Printf("Master playlist generated successfully: %s\n", masterPath)
	fmt.Printf("Content:\n%s\n", masterPlaylist)
}

// generateMasterPlaylistContent generates the master HLS playlist content
func generateMasterPlaylistContent(profiles []models.VideoProfile) string {
	var content strings.Builder
	content.WriteString("#EXTM3U\n")
	content.WriteString("#EXT-X-VERSION:3\n\n")

	for _, profile := range profiles {
		if profile.Status == "completed" {
			bandwidth := profile.Bitrate * 1000 // Convert to bits per second
			content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n", bandwidth, profile.Resolution))
			content.WriteString(fmt.Sprintf("%s/playlist.m3u8\n\n", profile.Resolution))
		}
	}

	return content.String()
}
