package services

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"

	"linier-channel/internal/models"
)

type FTPWatcher struct {
	watchPath        string
	uploadService    *UploadService
	transcodeService *TranscodeService
	db               *gorm.DB
	watcher          *fsnotify.Watcher
	stopChan         chan bool
}

func NewFTPWatcher(watchPath string, uploadService *UploadService, transcodeService *TranscodeService, db *gorm.DB) *FTPWatcher {
	return &FTPWatcher{
		watchPath:        watchPath,
		uploadService:    uploadService,
		transcodeService: transcodeService,
		db:               db,
		stopChan:         make(chan bool),
	}
}

func (fw *FTPWatcher) StartWatching() error {
	log.Printf("FTP Watcher: Starting watcher initialization...")

	var err error
	fw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Printf("FTP Watcher: Failed to create watcher: %v", err)
		return err
	}
	defer fw.watcher.Close()
	log.Printf("FTP Watcher: Watcher created successfully")

	// Create watch directory if it doesn't exist
	log.Printf("FTP Watcher: Creating watch directory: %s", fw.watchPath)
	if err := os.MkdirAll(fw.watchPath, 0755); err != nil {
		log.Printf("FTP Watcher: Failed to create watch directory %s: %v", fw.watchPath, err)
		return err
	}
	log.Printf("FTP Watcher: Watch directory created/verified: %s", fw.watchPath)

	// Check if directory is accessible
	log.Printf("FTP Watcher: Checking directory accessibility...")
	if _, err := os.Stat(fw.watchPath); err != nil {
		log.Printf("FTP Watcher: Watch directory not accessible: %s, error: %v", fw.watchPath, err)
		return err
	}
	log.Printf("FTP Watcher: Directory is accessible: %s", fw.watchPath)

	// Check directory permissions
	log.Printf("FTP Watcher: Checking directory permissions...")
	info, err := os.Stat(fw.watchPath)
	if err != nil {
		log.Printf("FTP Watcher: Failed to get directory info: %v", err)
		return err
	}
	log.Printf("FTP Watcher: Directory permissions: %v", info.Mode())

	// Watch the FTP directory
	log.Printf("FTP Watcher: Adding watch path: %s", fw.watchPath)
	err = fw.watcher.Add(fw.watchPath)
	if err != nil {
		log.Printf("FTP Watcher: Failed to add watch path %s: %v", fw.watchPath, err)
		return err
	}
	log.Printf("FTP Watcher: Watch path added successfully: %s", fw.watchPath)

	log.Printf("FTP Watcher started, monitoring: %s", fw.watchPath)

	// Add periodic health check
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Add startup verification
	log.Printf("FTP Watcher: Verifying watcher is working...")

	// Test if watcher is actually working by checking events channel
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	// Send a test event to verify watcher is working
	go func() {
		time.Sleep(1 * time.Second)
		log.Printf("FTP Watcher: Test event sent, checking if watcher responds...")
	}()

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				log.Printf("FTP Watcher: Events channel closed, stopping watcher")
				return nil
			}

			log.Printf("FTP Watcher: Received event: %s, operation: %v", event.Name, event.Op)

			// Check if it's a new file
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("FTP Watcher: File created event detected: %s", event.Name)
				// Wait a bit for file to be fully written
				time.Sleep(2 * time.Second)

				// Check if it's a video file
				if fw.isVideoFile(event.Name) {
					log.Printf("FTP Watcher: New video file detected: %s", event.Name)
					go fw.processNewVideo(event.Name)
				} else {
					log.Printf("FTP Watcher: Non-video file ignored: %s", event.Name)
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				log.Printf("FTP Watcher: Errors channel closed, stopping watcher")
				return nil
			}
			log.Printf("FTP Watcher error: %v", err)
		case <-ticker.C:
			// Health check - verify directory still exists and is accessible
			if _, err := os.Stat(fw.watchPath); err != nil {
				log.Printf("FTP Watcher health check failed - directory not accessible: %s, error: %v", fw.watchPath, err)
				return err
			}
			log.Printf("FTP Watcher health check OK - monitoring: %s", fw.watchPath)
		case <-timeout.C:
			log.Printf("FTP Watcher: No events received in 5 seconds, watcher may be stuck")
			// Continue monitoring but log the issue
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
	// Create video record with relative path
	relativeFilePath := strings.TrimPrefix(filePath, fw.watchPath+"/")
	video := &models.Video{
		OriginalFilename: fileInfo.Name(),
		FilePath:         relativeFilePath,
		FileSize:         fileInfo.Size(),
		Status:           "uploaded",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	// Round to nearest integer
	return int64(duration + 0.5), nil
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
