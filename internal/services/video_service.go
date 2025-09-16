package services

import (
	"linier-channel/internal/models"
	"time"

	"gorm.io/gorm"
)

type VideoService struct {
	db *gorm.DB
}

func NewVideoService(db *gorm.DB) *VideoService {
	return &VideoService{db: db}
}

// CreateVideo creates a new video record
func (s *VideoService) CreateVideo(filename, filePath string, fileSize int64, duration int) (*models.Video, error) {
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

// GetVideoByID retrieves a video by ID
func (s *VideoService) GetVideoByID(id uint) (*models.Video, error) {
	var video models.Video
	if err := s.db.Preload("VideoProfiles").First(&video, id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

// GetVideoStatus returns the status of a video and its profiles
func (s *VideoService) GetVideoStatus(id uint) (*models.VideoStatusResponse, error) {
	var video models.Video
	if err := s.db.Preload("VideoProfiles").First(&video, id).Error; err != nil {
		return nil, err
	}

	// Calculate overall progress
	totalProfiles := len(video.VideoProfiles)
	completedProfiles := 0
	overallProgress := 0

	var profiles []models.VideoProfileStatus
	for _, profile := range video.VideoProfiles {
		if profile.Status == "completed" {
			completedProfiles++
		}
		overallProgress += profile.ProgressPercentage

		profiles = append(profiles, models.VideoProfileStatus{
			ProfileID:  profile.ID,
			Resolution: profile.Resolution,
			Status:     profile.Status,
			Progress:   profile.ProgressPercentage,
			Error:      profile.ErrorMessage,
		})
	}

	if totalProfiles > 0 {
		overallProgress = overallProgress / totalProfiles
	}

	// Determine overall status
	overallStatus := video.Status
	if completedProfiles == totalProfiles && totalProfiles > 0 {
		overallStatus = "completed"
	} else if completedProfiles > 0 {
		overallStatus = "processing"
	}

	return &models.VideoStatusResponse{
		VideoID:   video.ID,
		Status:    overallStatus,
		Progress:  overallProgress,
		Profiles:  profiles,
		CreatedAt: video.CreatedAt,
	}, nil
}

// UpdateVideoStatus updates the video status
func (s *VideoService) UpdateVideoStatus(id uint, status string, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}

	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	return s.db.Model(&models.Video{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateVideoProfileStatus updates a specific video profile status
func (s *VideoService) UpdateVideoProfileStatus(profileID uint, status string, progress int, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
		"progress_percentage": progress,
	}

	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	if status == "processing" {
		now := time.Now()
		updates["started_at"] = &now
	} else if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.Model(&models.VideoProfile{}).Where("id = ?", profileID).Updates(updates).Error
}

// GetVideosByStatus retrieves videos by status
func (s *VideoService) GetVideosByStatus(status string) ([]models.Video, error) {
	var videos []models.Video
	query := s.db.Preload("VideoProfiles")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

// GetVideos retrieves all videos with pagination
func (s *VideoService) GetVideos(limit, offset int) ([]models.Video, error) {
	var videos []models.Video
	if err := s.db.Preload("VideoProfiles").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// DeleteVideo deletes a video and all related data
func (s *VideoService) DeleteVideo(id uint) error {
	return s.db.Delete(&models.Video{}, id).Error
}
