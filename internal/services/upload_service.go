package services

import (
	"fmt"
	"io"
	"linier-channel/internal/config"
	"linier-channel/internal/models"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UploadService struct {
	db     *gorm.DB
	config *config.Config
}

func NewUploadService(db *gorm.DB, cfg *config.Config) *UploadService {
	return &UploadService{
		db:     db,
		config: cfg,
	}
}

// UploadVideo handles video file upload
func (s *UploadService) UploadVideo(file *multipart.FileHeader) (*models.UploadVideoResponse, error) {
	// Validate file
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Create upload directory if not exists
	uploadDir := s.config.Storage.UploadPath
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Get video duration using FFprobe
	duration, err := s.getVideoDuration(filePath)
	if err != nil {
		logrus.Warnf("Failed to get video duration: %v", err)
		duration = 0
	}

	// Create video record in database
	video, err := s.createVideoRecord(file.Filename, filePath, fileInfo.Size(), duration)
	if err != nil {
		// Clean up uploaded file if database operation fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create video record: %w", err)
	}

	return &models.UploadVideoResponse{
		VideoID: video.ID,
		Message: "Video uploaded successfully",
	}, nil
}

// validateFile validates the uploaded file
func (s *UploadService) validateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > s.config.Storage.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.config.Storage.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedFormats := s.config.Storage.AllowedFormats
	
	validFormat := false
	for _, format := range allowedFormats {
		if ext == "."+format {
			validFormat = true
			break
		}
	}

	if !validFormat {
		return fmt.Errorf("unsupported file format. Allowed formats: %v", allowedFormats)
	}

	return nil
}

// getVideoDuration gets video duration using FFprobe
func (s *UploadService) getVideoDuration(filePath string) (int, error) {
	cmd := exec.Command(s.config.FFmpeg.FFprobePath, 
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
	
	return int(duration), nil
}

// createVideoRecord creates a video record in the database
func (s *UploadService) createVideoRecord(filename, filePath string, fileSize int64, duration int) (*models.Video, error) {
	video := &models.Video{
		OriginalFilename: filename,
		FilePath:         filePath,
		FileSize:         fileSize,
		Duration:         duration,
		Status:           "uploaded",
	}

	if err := s.db.Create(video).Error; err != nil {
		return nil, err
	}

	// Create video profiles for the video
	profiles := []string{"720p", "480p", "360p"}
	bitrates := []int{2000, 1000, 500}

	for i, profile := range profiles {
		videoProfile := &models.VideoProfile{
			VideoID:      video.ID,
			Resolution:   profile,
			Bitrate:      bitrates[i],
			AudioBitrate: 128,
			SegmentTime:  4,
			Status:       "pending",
		}

		if err := s.db.Create(videoProfile).Error; err != nil {
			return nil, err
		}
	}

	return video, nil
}

// GetUploadedVideos retrieves uploaded videos
func (s *UploadService) GetUploadedVideos() ([]models.Video, error) {
	var videos []models.Video
	if err := s.db.Preload("VideoProfiles").
		Where("status = ?", "uploaded").
		Order("created_at DESC").
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// DeleteUploadedVideo deletes an uploaded video and its file
func (s *UploadService) DeleteUploadedVideo(id uint) error {
	var video models.Video
	if err := s.db.First(&video, id).Error; err != nil {
		return err
	}

	// Delete file from filesystem
	if err := os.Remove(video.FilePath); err != nil {
		logrus.Warnf("Failed to delete file %s: %v", video.FilePath, err)
	}

	// Delete from database (cascade will handle related records)
	return s.db.Delete(&video).Error
}
