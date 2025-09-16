package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"linier-channel/internal/config"
	"linier-channel/internal/models"
	"linier-channel/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type TranscodeService struct {
	db     *gorm.DB
	config *config.Config
	redis  *redis.Client
}

func NewTranscodeService(db *gorm.DB, cfg *config.Config) *TranscodeService {
	return &TranscodeService{
		db:     db,
		config: cfg,
	}
}


// SetRedisClient sets the Redis client for the service
func (s *TranscodeService) SetRedisClient(redis *redis.Client) {
	s.redis = redis
}

// QueueTranscodeJob queues a transcoding job
func (s *TranscodeService) QueueTranscodeJob(videoID, profileID uint, priority int) error {
	job := &models.TranscodeJob{
		VideoID:   videoID,
		ProfileID: profileID,
		Status:    "queued",
		Priority:  priority,
	}

	if err := s.db.Create(job).Error; err != nil {
		return err
	}

	// Add to Redis queue
	if s.redis != nil {
		ctx := context.Background()
		jobData := fmt.Sprintf("%d:%d", videoID, profileID)
		return s.redis.LPush(ctx, "transcode_queue", jobData).Err()
	}

	return nil
}

// ProcessTranscodeJob processes a transcoding job
func (s *TranscodeService) ProcessTranscodeJob(videoID, profileID uint) error {
	// Get video and profile information
	var video models.Video
	if err := s.db.Preload("VideoProfiles").First(&video, videoID).Error; err != nil {
		return err
	}

	var profile models.VideoProfile
	if err := s.db.First(&profile, profileID).Error; err != nil {
		return err
	}

	// Update job status to processing
	s.updateJobStatus(videoID, profileID, "processing", "")

	// Update profile status to processing
	s.updateProfileStatus(profileID, "processing", 0, "")

	// Create output directory with date structure
	outputDir := utils.GenerateTranscodedPath(s.config.Storage.TranscodedPath, video.OriginalFilename, profile.Resolution)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate FFmpeg command
	cmd, err := s.generateFFmpegCommand(video.FilePath, outputDir, &profile)
	if err != nil {
		s.updateJobStatus(videoID, profileID, "failed", err.Error())
		s.updateProfileStatus(profileID, "failed", 0, err.Error())
		return err
	}

	// Execute FFmpeg command
	if err := s.executeFFmpegCommand(cmd, uint(profileID)); err != nil {
		s.updateJobStatus(videoID, profileID, "failed", err.Error())
		s.updateProfileStatus(profileID, "failed", 0, err.Error())
		return err
	}

	// Update job and profile status to completed
	s.updateJobStatus(videoID, profileID, "completed", "")
	s.updateProfileStatus(profileID, "completed", 100, "")

	// Update playlist path
	playlistPath := filepath.Join(outputDir, "playlist.m3u8")
	s.updatePlaylistPath(profileID, playlistPath)

	// Check if all profiles are completed
	s.checkVideoCompletion(videoID)

	return nil
}

// generateFFmpegCommand generates FFmpeg command for transcoding
func (s *TranscodeService) generateFFmpegCommand(inputPath, outputDir string, profile *models.VideoProfile) (*exec.Cmd, error) {
	// Get video dimensions based on resolution
	width, height := s.getVideoDimensions(profile.Resolution)
	
	// Generate output path
	outputPath := filepath.Join(outputDir, "playlist.m3u8")

	// Build FFmpeg command
	args := []string{
		"-i", inputPath,
		"-c:v", profile.CodecVideo,
		"-c:a", profile.CodecAudio,
		"-b:v", fmt.Sprintf("%dk", profile.Bitrate),
		"-b:a", fmt.Sprintf("%dk", profile.AudioBitrate),
		"-s", fmt.Sprintf("%dx%d", width, height),
		"-f", "hls",
		"-hls_time", strconv.Itoa(profile.SegmentTime),
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outputDir, "%03d.ts"),
		"-hls_flags", "independent_segments+split_by_time",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", profile.SegmentTime),
		"-segment_time_metadata", "1",
		outputPath,
	}

	cmd := exec.Command(s.config.FFmpeg.FFmpegPath, args...)
	return cmd, nil
}

// getVideoDimensions returns video dimensions based on resolution
func (s *TranscodeService) getVideoDimensions(resolution string) (int, int) {
	switch resolution {
	case "720p":
		return 1280, 720
	case "480p":
		return 854, 480
	case "360p":
		return 640, 360
	default:
		return 640, 360
	}
}

// executeFFmpegCommand executes the FFmpeg command
func (s *TranscodeService) executeFFmpegCommand(cmd *exec.Cmd, profileID uint) error {
	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return err
	}

	// Update total segments count
	outputDir := filepath.Dir(cmd.Args[len(cmd.Args)-1])
	segmentCount := s.countSegments(outputDir)
	s.updateSegmentCount(uint(profileID), segmentCount)

	return nil
}

// countSegments counts the number of .ts segments in a directory
func (s *TranscodeService) countSegments(dir string) int {
	files, err := filepath.Glob(filepath.Join(dir, "*.ts"))
	if err != nil {
		return 0
	}
	return len(files)
}

// updateJobStatus updates the job status
func (s *TranscodeService) updateJobStatus(videoID, profileID uint, status, errorMessage string) {
	updates := map[string]interface{}{
		"status": status,
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

	s.db.Model(&models.TranscodeJob{}).
		Where("video_id = ? AND profile_id = ?", videoID, profileID).
		Updates(updates)
}

// updateProfileStatus updates the profile status
func (s *TranscodeService) updateProfileStatus(profileID uint, status string, progress int, errorMessage string) {
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

	s.db.Model(&models.VideoProfile{}).Where("id = ?", profileID).Updates(updates)
}

// updatePlaylistPath updates the playlist path
func (s *TranscodeService) updatePlaylistPath(profileID uint, playlistPath string) {
	s.db.Model(&models.VideoProfile{}).
		Where("id = ?", profileID).
		Update("playlist_path", playlistPath)
}

// updateSegmentCount updates the total segments count
func (s *TranscodeService) updateSegmentCount(profileID uint, count int) {
	s.db.Model(&models.VideoProfile{}).
		Where("id = ?", profileID).
		Update("total_segments", count)
}

// checkVideoCompletion checks if all profiles are completed
func (s *TranscodeService) checkVideoCompletion(videoID uint) {
	var profiles []models.VideoProfile
	s.db.Where("video_id = ?", videoID).Find(&profiles)

	allCompleted := true
	for _, profile := range profiles {
		if profile.Status != "completed" {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		// Generate and save master playlist
		masterPath := s.generateAndSaveMasterPlaylist(videoID)
		
		// Update video status to completed and set video path
		s.db.Model(&models.Video{}).
			Where("id = ?", videoID).
			Updates(map[string]interface{}{
				"status": "completed",
				"video_path": masterPath,
			})
		
		// Archive original file
		s.archiveOriginalFile(videoID)
	}
}

// generateAndSaveMasterPlaylist generates and saves the master playlist
func (s *TranscodeService) generateAndSaveMasterPlaylist(videoID uint) string {
	// Get video and profiles
	var video models.Video
	if err := s.db.Preload("VideoProfiles").First(&video, videoID).Error; err != nil {
		return ""
	}

	// Generate master playlist content
	masterPlaylist := s.generateMasterPlaylistContent(video.VideoProfiles)
	
	// Save master playlist to file using the same path structure as transcoded files
	// Use the same base directory as the first profile
	if len(video.VideoProfiles) == 0 {
		return ""
	}
	
	// Get the base directory from the first profile's playlist path
	firstProfile := video.VideoProfiles[0]
	if firstProfile.PlaylistPath == "" {
		return ""
	}
	
	// Extract base directory from playlist path (remove resolution and playlist.m3u8)
	baseDir := filepath.Dir(filepath.Dir(firstProfile.PlaylistPath))
	masterPath := filepath.Join(s.config.Storage.TranscodedPath, baseDir, "master.m3u8")
	
	if err := os.MkdirAll(filepath.Dir(masterPath), 0755); err != nil {
		return ""
	}

	ioutil.WriteFile(masterPath, []byte(masterPlaylist), 0644)
	
	// Return relative path for database storage
	relativePath := strings.TrimPrefix(masterPath, s.config.Storage.TranscodedPath+"/")
	return relativePath
}

// generateMasterPlaylistContent generates the master HLS playlist content
func (s *TranscodeService) generateMasterPlaylistContent(profiles []models.VideoProfile) string {
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

// archiveOriginalFile moves the original file to archive directory
func (s *TranscodeService) archiveOriginalFile(videoID uint) {
	var video models.Video
	if err := s.db.First(&video, videoID).Error; err != nil {
		return // Silently fail if video not found
	}

	// Generate archive path
	archiveDir := utils.GenerateArchivePath(s.config.Storage.ArchivePath, video.OriginalFilename)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return // Silently fail if can't create archive directory
	}

	// Move original file to archive
	originalPath := video.FilePath
	archivePath := filepath.Join(archiveDir, video.OriginalFilename)
	
	if err := os.Rename(originalPath, archivePath); err != nil {
		return // Silently fail if can't move file
	}

	// Update file path in database
	s.db.Model(&models.Video{}).
		Where("id = ?", videoID).
		Update("file_path", archivePath)
}

// GetTranscodeQueue returns the current transcode queue
func (s *TranscodeService) GetTranscodeQueue() ([]models.TranscodeJob, error) {
	var jobs []models.TranscodeJob
	if err := s.db.Preload("Video").Preload("Profile").
		Where("status = ?", "queued").
		Order("priority DESC, created_at ASC").
		Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetTranscodeStatus returns the status of transcoding jobs
func (s *TranscodeService) GetTranscodeStatus() (map[string]int, error) {
	status := make(map[string]int)

	var counts []struct {
		Status string
		Count  int
	}

	if err := s.db.Model(&models.TranscodeJob{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, err
	}

	for _, count := range counts {
		status[count.Status] = count.Count
	}

	return status, nil
}
