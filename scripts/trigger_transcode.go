package main

import (
	"fmt"
	"log"
	"linier-channel/internal/config"
	"linier-channel/internal/database"
	"linier-channel/internal/services"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run scripts/trigger_transcode.go <video_id>")
	}

	videoID, err := strconv.ParseUint(os.Args[1], 10, 32)
	if err != nil {
		log.Fatal("Invalid video ID")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	redisClient, err := database.InitializeRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Initialize transcode service
	transcodeService := services.NewTranscodeService(db, cfg)
	transcodeService.SetRedisClient(redisClient)

	// Get video profiles for this video
	var profiles []struct {
		ID uint `json:"id"`
	}
	
	if err := db.Model(&struct{ ID uint }{}).
		Table("video_profiles").
		Where("video_id = ? AND status = ?", videoID, "pending").
		Find(&profiles).Error; err != nil {
		log.Fatalf("Failed to get video profiles: %v", err)
	}

	// Queue transcoding jobs
	for _, profile := range profiles {
		if err := transcodeService.QueueTranscodeJob(uint(videoID), profile.ID, 1); err != nil {
			log.Printf("Failed to queue job for profile %d: %v", profile.ID, err)
		} else {
			fmt.Printf("Queued transcoding job for video %d, profile %d\n", videoID, profile.ID)
		}
	}

	fmt.Printf("Successfully queued %d transcoding jobs for video %d\n", len(profiles), videoID)
}
