package main

import (
	"linier-channel/internal/config"
	"linier-channel/internal/database"
	"linier-channel/internal/handlers"
	"linier-channel/internal/services"
	"linier-channel/internal/worker"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	setupLogging(cfg.Logging)

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

	// Initialize services
	videoService := services.NewVideoService(db)
	transcodeService := services.NewTranscodeService(db, cfg)
	transcodeService.SetRedisClient(redisClient) // Set Redis client for transcoding
	playlistService := services.NewPlaylistService(db, cfg)
	uploadService := services.NewUploadService(db, cfg)
	uploadService.SetTranscodeService(transcodeService) // Set transcode service for upload service

	// Initialize handlers
	handlers := handlers.NewHandlers(videoService, transcodeService, playlistService, uploadService)

	// Start transcode workers
	workerManager := worker.NewWorkerManager(transcodeService, redisClient, cfg)
	go workerManager.Start()

	// Initialize FTP Watcher
	ftpWatcher := services.NewFTPWatcher(cfg.Storage.UploadPath, uploadService, transcodeService, db)
	go func() {
		if err := ftpWatcher.StartWatching(); err != nil {
			log.Printf("FTP Watcher error: %v", err)
		}
	}()

	// Start HTTP server
	server := handlers.SetupRoutes()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		ftpWatcher.Stop()
		workerManager.Stop()
		os.Exit(0)
	}()

	log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
	if err := server.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupLogging configures logging based on the configuration
func setupLogging(logConfig config.LoggingConfig) {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logConfig.Path)
	if logDir != "." {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("Warning: Failed to create log directory %s: %v", logDir, err)
			return
		}
	}

	// Open log file
	logFile, err := os.OpenFile(logConfig.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: Failed to open log file %s: %v", logConfig.Path, err)
		return
	}

	// Set log output to file
	log.SetOutput(logFile)
	log.Printf("Logging configured - Level: %s, Format: %s, Path: %s", logConfig.Level, logConfig.Format, logConfig.Path)
}
