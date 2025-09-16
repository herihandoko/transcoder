package services

import (
	"fmt"
	"io/ioutil"
	"linier-channel/internal/config"
	"linier-channel/internal/models"
	"linier-channel/internal/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PlaylistService struct {
	db     *gorm.DB
	config *config.Config
}

func NewPlaylistService(db *gorm.DB, cfg *config.Config) *PlaylistService {
	return &PlaylistService{db: db, config: cfg}
}


// CreatePlaylist creates a new playlist
func (s *PlaylistService) CreatePlaylist(req *models.CreatePlaylistRequest) (*models.Playlist, error) {
	playlist := &models.Playlist{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.db.Create(playlist).Error; err != nil {
		return nil, err
	}

	// Add videos to playlist if provided
	if len(req.VideoIDs) > 0 {
		for i, videoID := range req.VideoIDs {
		playlistVideo := &models.PlaylistVideo{
			PlaylistID: playlist.ID,
			VideoID:    uint(videoID),
				SortOrder:  i + 1,
			}
			s.db.Create(playlistVideo)
		}
	}

	return playlist, nil
}

// GetPlaylist retrieves a playlist by ID
func (s *PlaylistService) GetPlaylist(id uint) (*models.PlaylistResponse, error) {
	var playlist models.Playlist
	if err := s.db.Preload("PlaylistVideos.Video.VideoProfiles").
		First(&playlist, id).Error; err != nil {
		return nil, err
	}

	// Build video info list
	var videos []models.VideoInfo
	for _, pv := range playlist.PlaylistVideos {
		video := pv.Video
		videos = append(videos, models.VideoInfo{
			VideoID:   uint(video.ID),
			Filename:  video.OriginalFilename,
			Duration:  video.Duration,
			Status:    video.Status,
			SortOrder: pv.SortOrder,
		})
	}

	// Sort by sort order
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].SortOrder < videos[j].SortOrder
	})

	return &models.PlaylistResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsActive:    playlist.IsActive,
		Videos:      videos,
		CreatedAt:   playlist.CreatedAt,
	}, nil
}

// GetPlaylists retrieves all playlists
func (s *PlaylistService) GetPlaylists() ([]models.PlaylistResponse, error) {
	var playlists []models.Playlist
	if err := s.db.Preload("PlaylistVideos.Video").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&playlists).Error; err != nil {
		return nil, err
	}

	var responses []models.PlaylistResponse
	for _, playlist := range playlists {
		var videos []models.VideoInfo
		for _, pv := range playlist.PlaylistVideos {
			video := pv.Video
			videos = append(videos, models.VideoInfo{
				VideoID:   video.ID,
				Filename:  video.OriginalFilename,
				Duration:  video.Duration,
				Status:    video.Status,
				SortOrder: pv.SortOrder,
			})
		}

		// Sort by sort order
		sort.Slice(videos, func(i, j int) bool {
			return videos[i].SortOrder < videos[j].SortOrder
		})

		responses = append(responses, models.PlaylistResponse{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsActive:    playlist.IsActive,
			Videos:      videos,
			CreatedAt:   playlist.CreatedAt,
		})
	}

	return responses, nil
}

// AddVideoToPlaylist adds a video to a playlist
func (s *PlaylistService) AddVideoToPlaylist(playlistID uint, req *models.AddVideoToPlaylistRequest) error {
	// Check if video already exists in playlist
	var existing models.PlaylistVideo
	if err := s.db.Where("playlist_id = ? AND video_id = ?", playlistID, req.VideoID).
		First(&existing).Error; err == nil {
		return fmt.Errorf("video already exists in playlist")
	}

	// Get next sort order if not provided
	sortOrder := req.SortOrder
	if sortOrder == 0 {
		var maxOrder int
		s.db.Model(&models.PlaylistVideo{}).
			Where("playlist_id = ?", playlistID).
			Select("COALESCE(MAX(sort_order), 0)").
			Scan(&maxOrder)
		sortOrder = maxOrder + 1
	}

	playlistVideo := &models.PlaylistVideo{
		PlaylistID: playlistID,
		VideoID:    req.VideoID,
		SortOrder:  sortOrder,
	}

	return s.db.Create(playlistVideo).Error
}

// RemoveVideoFromPlaylist removes a video from a playlist
func (s *PlaylistService) RemoveVideoFromPlaylist(playlistID, videoID uint) error {
	return s.db.Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
		Delete(&models.PlaylistVideo{}).Error
}

// GenerateHLSPlaylist generates HLS playlist for a video
func (s *PlaylistService) GenerateHLSPlaylist(videoID uint) (*models.HLSPlaylistResponse, error) {
	var video models.Video
	if err := s.db.Preload("VideoProfiles").
		First(&video, videoID).Error; err != nil {
		return nil, err
	}

	// Check if video is completed
	if video.Status != "completed" {
		return nil, fmt.Errorf("video is not completed yet")
	}

	// Generate master playlist
	masterPlaylist := s.generateMasterPlaylist(videoID, video.VideoProfiles)

	// Build profiles info
	var profiles []models.HLSProfile
	for _, profile := range video.VideoProfiles {
		if profile.Status == "completed" {
			profiles = append(profiles, models.HLSProfile{
				Resolution:  profile.Resolution,
				Bitrate:     profile.Bitrate,
				PlaylistURL: fmt.Sprintf("/videos/%d/%s/playlist.m3u8", videoID, profile.Resolution),
				Status:      profile.Status,
			})
		}
	}

	return &models.HLSPlaylistResponse{
		MasterPlaylist: masterPlaylist,
		Profiles:       profiles,
	}, nil
}

// generateMasterPlaylist generates the master HLS playlist content
func (s *PlaylistService) generateMasterPlaylist(videoID uint, profiles []models.VideoProfile) string {
	var lines []string
	lines = append(lines, "#EXTM3U")
	lines = append(lines, "#EXT-X-VERSION:3")

	for _, profile := range profiles {
		if profile.Status == "completed" {
			bandwidth := profile.Bitrate * 1000 // Convert to bits per second
			width, height := s.getVideoDimensions(profile.Resolution)
			
			line := fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d",
				bandwidth, width, height)
			lines = append(lines, line)
			lines = append(lines, fmt.Sprintf("%s/playlist.m3u8", profile.Resolution))
		}
	}

	return strings.Join(lines, "\n")
}

// getVideoDimensions returns video dimensions based on resolution
func (s *PlaylistService) getVideoDimensions(resolution string) (int, int) {
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

// SaveMasterPlaylist saves the master playlist to file
func (s *PlaylistService) SaveMasterPlaylist(videoID uint, content string) error {
	// Get video filename for directory name
	var video models.Video
	if err := s.db.First(&video, videoID).Error; err != nil {
		return fmt.Errorf("failed to get video: %v", err)
	}
	
	// Generate path with date structure (without resolution for master playlist)
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	dirName := utils.SanitizeFilename(video.OriginalFilename)
	videoDir := filepath.Join(s.config.Storage.TranscodedPath, year, month, dirName)
	
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return err
	}

	// Save master playlist
	masterPath := filepath.Join(videoDir, "master.m3u8")
	return ioutil.WriteFile(masterPath, []byte(content), 0644)
}

// GetPlaylistFile returns the content of a playlist file
func (s *PlaylistService) GetPlaylistFile(videoID uint, resolution string) (string, error) {
	// Get video filename for directory name
	var video models.Video
	if err := s.db.First(&video, videoID).Error; err != nil {
		return "", fmt.Errorf("failed to get video: %v", err)
	}
	
	// Generate path with date structure
	playlistPath := utils.GenerateTranscodedPath(s.config.Storage.TranscodedPath, video.OriginalFilename, resolution)
	playlistPath = filepath.Join(playlistPath, "playlist.m3u8")
	
	content, err := ioutil.ReadFile(playlistPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GetSegmentFile returns the content of a segment file
func (s *PlaylistService) GetSegmentFile(videoID uint, resolution, segment string) ([]byte, error) {
	// Get video filename for directory name
	var video models.Video
	if err := s.db.First(&video, videoID).Error; err != nil {
		return nil, fmt.Errorf("failed to get video: %v", err)
	}
	
	// Generate path with date structure
	segmentPath := utils.GenerateTranscodedPath(s.config.Storage.TranscodedPath, video.OriginalFilename, resolution)
	segmentPath = filepath.Join(segmentPath, segment)
	return ioutil.ReadFile(segmentPath)
}

// DeletePlaylist deletes a playlist
func (s *PlaylistService) DeletePlaylist(id uint) error {
	return s.db.Delete(&models.Playlist{}, id).Error
}

// UpdatePlaylist updates a playlist
func (s *PlaylistService) UpdatePlaylist(id uint, name, description string, isActive bool) error {
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"is_active":   isActive,
	}

	return s.db.Model(&models.Playlist{}).Where("id = ?", id).Updates(updates).Error
}
