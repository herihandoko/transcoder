package main

import (
	"log"
	"linier-channel/internal/config"
	"linier-channel/internal/database"
	"linier-channel/internal/handlers"
	"linier-channel/internal/services"
	"linier-channel/internal/worker"
	"os"
	"os/signal"
	"syscall"
)

func main() {
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

	// Initialize services
	videoService := services.NewVideoService(db)
	transcodeService := services.NewTranscodeService(db, cfg)
	playlistService := services.NewPlaylistService(db, cfg)
	uploadService := services.NewUploadService(db, cfg)

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
