package services

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"

	"linier-channel/internal/models"
)

type FTPWatcher struct {
	watchPath       string
	uploadService   *UploadService
	transcodeService *TranscodeService
	db             *gorm.DB
	watcher        *fsnotify.Watcher
	stopChan       chan bool
}

func NewFTPWatcher(watchPath string, uploadService *UploadService, transcodeService *TranscodeService, db *gorm.DB) *FTPWatcher {
	return &FTPWatcher{
		watchPath:       watchPath,
		uploadService:   uploadService,
		transcodeService: transcodeService,
		db:             db,
		stopChan:       make(chan bool),
	}
}

func (fw *FTPWatcher) StartWatching() error {
	var err error
	fw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fw.watcher.Close()

	// Create watch directory if it doesn't exist
	if err := os.MkdirAll(fw.watchPath, 0755); err != nil {
		return err
	}

	// Watch the FTP directory
	err = fw.watcher.Add(fw.watchPath)
	if err != nil {
		return err
	}

	log.Printf("FTP Watcher started, monitoring: %s", fw.watchPath)

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}
			
			// Check if it's a new file
			if event.Op&fsnotify.Create == fsnotify.Create {
				// Wait a bit for file to be fully written
				time.Sleep(2 * time.Second)
				
				// Check if it's a video file
				if fw.isVideoFile(event.Name) {
					log.Printf("New video file detected: %s", event.Name)
					go fw.processNewVideo(event.Name)
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("FTP Watcher error: %v", err)
		case <-fw.stopChan:
			log.Println("FTP Watcher stopping...")
			return nil
		}
	}
}

func (fw *FTPWatcher) Stop() {
	close(fw.stopChan)
}

func (fw *FTPWatcher) isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}
	
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

func (fw *FTPWatcher) processNewVideo(filePath string) {
	// Check if file exists and is readable
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File does not exist: %s", filePath)
		return
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to get file info: %v", err)
		return
	}

	// Check if file is not empty
	if fileInfo.Size() == 0 {
		log.Printf("File is empty: %s", filePath)
		return
	}

	// Create video record in database
	video, err := fw.createVideoFromFile(filePath, fileInfo)
	if err != nil {
		log.Printf("Failed to create video record: %v", err)
		return
	}

	// Queue transcoding jobs
	err = fw.queueTranscodingJobs(video.ID)
	if err != nil {
		log.Printf("Failed to queue transcoding jobs: %v", err)
		return
	}

	log.Printf("Video %d queued for transcoding from FTP file: %s", video.ID, filePath)
}

func (fw *FTPWatcher) createVideoFromFile(filePath string, fileInfo os.FileInfo) (*models.Video, error) {
	// Create video record
	video := &models.Video{
		OriginalFilename: fileInfo.Name(),
		FilePath:        filePath,
		FileSize:        fileInfo.Size(),
		Status:          "uploaded",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Get video duration using ffprobe
	duration, err := fw.getVideoDuration(filePath)
	if err != nil {
		log.Printf("Failed to get video duration: %v", err)
		duration = 0
	}
	video.Duration = int(duration)

	// Save to database
	if err := fw.db.Create(video).Error; err != nil {
		return nil, err
	}

	// Create video profiles
	profiles := []models.VideoProfile{
		{
			VideoID:      video.ID,
			Resolution:   "720p",
			CodecVideo:   "h264",
			CodecAudio:   "aac",
			Bitrate:      2000,
			AudioBitrate: 128,
			SegmentTime:  4,
			Status:       "pending",
			CreatedAt:    time.Now(),
		},
		{
			VideoID:      video.ID,
			Resolution:   "480p",
			CodecVideo:   "h264",
			CodecAudio:   "aac",
			Bitrate:      1000,
			AudioBitrate: 128,
			SegmentTime:  4,
			Status:       "pending",
			CreatedAt:    time.Now(),
		},
		{
			VideoID:      video.ID,
			Resolution:   "360p",
			CodecVideo:   "h264",
			CodecAudio:   "aac",
			Bitrate:      500,
			AudioBitrate: 128,
			SegmentTime:  4,
			Status:       "pending",
			CreatedAt:    time.Now(),
		},
	}

	for _, profile := range profiles {
		if err := fw.db.Create(&profile).Error; err != nil {
			log.Printf("Failed to create video profile: %v", err)
		}
	}

	return video, nil
}

func (fw *FTPWatcher) getVideoDuration(filePath string) (int64, error) {
	// Use ffprobe to get video duration
	// This is a simplified version - you might want to use the same logic as in upload_service.go
	return 0, nil // Placeholder - implement actual ffprobe logic
}

func (fw *FTPWatcher) queueTranscodingJobs(videoID uint) error {
	// Get all pending profiles for this video
	var profiles []models.VideoProfile
	if err := fw.db.Where("video_id = ? AND status = ?", videoID, "pending").Find(&profiles).Error; err != nil {
		return err
	}

	// Queue each profile for transcoding
	for _, profile := range profiles {
		err := fw.transcodeService.QueueTranscodeJob(videoID, profile.ID, 1)
		if err != nil {
			log.Printf("Failed to queue transcoding job for video %d, profile %d: %v", videoID, profile.ID, err)
		}
	}

	return nil
}
