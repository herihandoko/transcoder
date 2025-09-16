package models

import (
	"time"
)

// Video represents the main video entity
type Video struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	OriginalFilename string    `json:"original_filename" gorm:"size:255;not null"`
	FilePath         string    `json:"file_path" gorm:"size:500;not null"`
	FileSize         int64     `json:"file_size"`
	Duration         int       `json:"duration"` // in seconds
	Status           string    `json:"status" gorm:"type:enum('uploaded','processing','completed','failed');default:'uploaded'"`
	ErrorMessage     string    `json:"error_message" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	VideoProfiles []VideoProfile `json:"video_profiles" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
	PlaylistVideos []PlaylistVideo `json:"playlist_videos" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
}

// VideoProfile represents video transcoding profiles
type VideoProfile struct {
	ID               uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	VideoID          uint       `json:"video_id" gorm:"not null;index"`
	Resolution       string     `json:"resolution" gorm:"size:10;index"`
	CodecVideo       string     `json:"codec_video" gorm:"size:20;default:'h264'"`
	CodecAudio       string     `json:"codec_audio" gorm:"size:20;default:'aac'"`
	Bitrate          int        `json:"bitrate"` // in kbps
	AudioBitrate     int        `json:"audio_bitrate" gorm:"default:128"`
	SegmentTime      int        `json:"segment_time" gorm:"default:4"`
	TotalSegments    int        `json:"total_segments"`
	PlaylistPath     string     `json:"playlist_path" gorm:"size:500"`
	Status           string     `json:"status" gorm:"type:enum('pending','processing','completed','failed');default:'pending';index"`
	ProgressPercentage int      `json:"progress_percentage" gorm:"default:0"`
	ErrorMessage     string     `json:"error_message" gorm:"type:text"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatedAt        time.Time  `json:"created_at"`

	// Relationships
	Video Video `json:"video" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
}

// Playlist represents video playlists
type Playlist struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Relationships
	PlaylistVideos []PlaylistVideo `json:"playlist_videos" gorm:"foreignKey:PlaylistID;constraint:OnDelete:CASCADE"`
}

// PlaylistVideo represents the relationship between playlists and videos
type PlaylistVideo struct {
	PlaylistID uint `json:"playlist_id" gorm:"primaryKey"`
	VideoID    uint `json:"video_id" gorm:"primaryKey"`
	SortOrder  int    `json:"sort_order" gorm:"default:0;index"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Playlist Playlist `json:"playlist" gorm:"foreignKey:PlaylistID;constraint:OnDelete:CASCADE"`
	Video    Video    `json:"video" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
}

// TranscodeJob represents transcoding job queue
type TranscodeJob struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	VideoID     uint       `json:"video_id" gorm:"not null"`
	ProfileID   uint       `json:"profile_id" gorm:"not null"`
	Status      string     `json:"status" gorm:"type:enum('queued','processing','completed','failed');default:'queued';index"`
	Priority    int        `json:"priority" gorm:"default:0;index"`
	RetryCount  int        `json:"retry_count" gorm:"default:0"`
	MaxRetries  int        `json:"max_retries" gorm:"default:3"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"index"`

	// Relationships
	Video   Video        `json:"video" gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
	Profile VideoProfile `json:"profile" gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}

// SystemConfig represents system configuration
type SystemConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `json:"config_key" gorm:"size:100;uniqueIndex;not null"`
	ConfigValue string    `json:"config_value" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request/Response DTOs

// UploadVideoRequest represents video upload request
type UploadVideoRequest struct {
	Filename string `json:"filename" binding:"required"`
}

// UploadVideoResponse represents video upload response
type UploadVideoResponse struct {
	VideoID uint `json:"video_id"`
	Message string `json:"message"`
}

// VideoStatusResponse represents video status response
type VideoStatusResponse struct {
	VideoID   uint                  `json:"video_id"`
	Status    string                `json:"status"`
	Progress  int                   `json:"progress"`
	Profiles  []VideoProfileStatus  `json:"profiles"`
	CreatedAt time.Time             `json:"created_at"`
}

// VideoProfileStatus represents individual profile status
type VideoProfileStatus struct {
	ProfileID   uint `json:"profile_id"`
	Resolution  string `json:"resolution"`
	Status      string `json:"status"`
	Progress    int    `json:"progress_percentage"`
	Error       string `json:"error_message,omitempty"`
}

// CreatePlaylistRequest represents playlist creation request
type CreatePlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	VideoIDs    []uint `json:"video_ids"`
}

// AddVideoToPlaylistRequest represents add video to playlist request
type AddVideoToPlaylistRequest struct {
	VideoID   uint `json:"video_id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

// PlaylistResponse represents playlist response
type PlaylistResponse struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	IsActive    bool        `json:"is_active"`
	Videos      []VideoInfo `json:"videos"`
	CreatedAt   time.Time   `json:"created_at"`
}

// VideoInfo represents video information in playlist
type VideoInfo struct {
	VideoID   uint `json:"video_id"`
	Filename  string `json:"filename"`
	Duration  int    `json:"duration"`
	Status    string `json:"status"`
	SortOrder int    `json:"sort_order"`
}

// HLSPlaylistResponse represents HLS playlist response
type HLSPlaylistResponse struct {
	MasterPlaylist string `json:"master_playlist"`
	Profiles       []HLSProfile `json:"profiles"`
}

// HLSProfile represents HLS profile information
type HLSProfile struct {
	Resolution string `json:"resolution"`
	Bitrate    int    `json:"bitrate"`
	PlaylistURL string `json:"playlist_url"`
	Status     string `json:"status"`
}
