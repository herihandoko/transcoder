package handlers

import (
	"linier-channel/internal/models"
	"linier-channel/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	videoService     *services.VideoService
	transcodeService *services.TranscodeService
	playlistService  *services.PlaylistService
	uploadService    *services.UploadService
}

func NewHandlers(
	videoService *services.VideoService,
	transcodeService *services.TranscodeService,
	playlistService *services.PlaylistService,
	uploadService *services.UploadService,
) *Handlers {
	return &Handlers{
		videoService:     videoService,
		transcodeService: transcodeService,
		playlistService:  playlistService,
		uploadService:    uploadService,
	}
}

func (h *Handlers) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Video routes
		videos := v1.Group("/videos")
		{
			videos.POST("/upload", h.UploadVideo)
			videos.GET("/", h.GetVideos)
			videos.GET("/:id", h.GetVideo)
			videos.GET("/:id/status", h.GetVideoStatus)
			videos.DELETE("/:id", h.DeleteVideo)
		}

		// Playlist routes
		playlists := v1.Group("/playlists")
		{
			playlists.POST("/", h.CreatePlaylist)
			playlists.GET("/", h.GetPlaylists)
			playlists.GET("/:id", h.GetPlaylist)
			playlists.PUT("/:id", h.UpdatePlaylist)
			playlists.DELETE("/:id", h.DeletePlaylist)
			playlists.POST("/:id/videos", h.AddVideoToPlaylist)
			playlists.DELETE("/:id/videos/:videoId", h.RemoveVideoFromPlaylist)
		}

		// HLS streaming routes
		streaming := v1.Group("/stream")
		{
			streaming.GET("/:videoId/master.m3u8", h.GetMasterPlaylist)
			streaming.GET("/:videoId/:resolution/playlist.m3u8", h.GetPlaylistFile)
			streaming.GET("/:videoId/:resolution/:segment", h.GetSegment)
		}

		// Admin routes
		admin := v1.Group("/admin")
		{
			admin.GET("/transcode/queue", h.GetTranscodeQueue)
			admin.GET("/transcode/status", h.GetTranscodeStatus)
		}
	}

	return r
}

// Health check endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "linier-channel",
	})
}

// Upload video endpoint
func (h *Handlers) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	response, err := h.uploadService.UploadVideo(file)
	if err != nil {
		logrus.Errorf("Failed to upload video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Get videos endpoint
func (h *Handlers) GetVideos(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	videos, err := h.videoService.GetVideos(limit, offset)
	if err != nil {
		logrus.Errorf("Failed to get videos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

// Get video endpoint
func (h *Handlers) GetVideo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.videoService.GetVideoByID(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		logrus.Errorf("Failed to get video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// Get video status endpoint
func (h *Handlers) GetVideoStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	status, err := h.videoService.GetVideoStatus(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		logrus.Errorf("Failed to get video status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Delete video endpoint
func (h *Handlers) DeleteVideo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	if err := h.uploadService.DeleteUploadedVideo(uint(id)); err != nil {
		logrus.Errorf("Failed to delete video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
}

// Create playlist endpoint
func (h *Handlers) CreatePlaylist(c *gin.Context) {
	var req models.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.playlistService.CreatePlaylist(&req)
	if err != nil {
		logrus.Errorf("Failed to create playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

// Get playlists endpoint
func (h *Handlers) GetPlaylists(c *gin.Context) {
	playlists, err := h.playlistService.GetPlaylists()
	if err != nil {
		logrus.Errorf("Failed to get playlists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get playlists"})
		return
	}

	c.JSON(http.StatusOK, playlists)
}

// Get playlist endpoint
func (h *Handlers) GetPlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	playlist, err := h.playlistService.GetPlaylist(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
			return
		}
		logrus.Errorf("Failed to get playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get playlist"})
		return
	}

	c.JSON(http.StatusOK, playlist)
}

// Update playlist endpoint
func (h *Handlers) UpdatePlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.playlistService.UpdatePlaylist(uint(id), req.Name, req.Description, req.IsActive); err != nil {
		logrus.Errorf("Failed to update playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist updated successfully"})
}

// Delete playlist endpoint
func (h *Handlers) DeletePlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	if err := h.playlistService.DeletePlaylist(uint(id)); err != nil {
		logrus.Errorf("Failed to delete playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})
}

// Add video to playlist endpoint
func (h *Handlers) AddVideoToPlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	var req models.AddVideoToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.playlistService.AddVideoToPlaylist(uint(id), &req); err != nil {
		logrus.Errorf("Failed to add video to playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video added to playlist successfully"})
}

// Remove video from playlist endpoint
func (h *Handlers) RemoveVideoFromPlaylist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	videoIdStr := c.Param("videoId")
	videoId, err := strconv.ParseUint(videoIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	if err := h.playlistService.RemoveVideoFromPlaylist(uint(id), uint(videoId)); err != nil {
		logrus.Errorf("Failed to remove video from playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove video from playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video removed from playlist successfully"})
}

// Get master playlist endpoint
func (h *Handlers) GetMasterPlaylist(c *gin.Context) {
	videoIdStr := c.Param("videoId")
	videoId, err := strconv.ParseUint(videoIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	playlist, err := h.playlistService.GenerateHLSPlaylist(uint(videoId))
	if err != nil {
		logrus.Errorf("Failed to generate HLS playlist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate playlist"})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.String(http.StatusOK, playlist.MasterPlaylist)
}

// Get playlist file endpoint (for specific resolution)
func (h *Handlers) GetPlaylistFile(c *gin.Context) {
	videoIdStr := c.Param("videoId")
	videoId, err := strconv.ParseUint(videoIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	resolution := c.Param("resolution")
	playlist, err := h.playlistService.GetPlaylistFile(uint(videoId), resolution)
	if err != nil {
		logrus.Errorf("Failed to get playlist file: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.String(http.StatusOK, playlist)
}

// Get segment endpoint
func (h *Handlers) GetSegment(c *gin.Context) {
	videoIdStr := c.Param("videoId")
	videoId, err := strconv.ParseUint(videoIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	resolution := c.Param("resolution")
	segment := c.Param("segment")

	segmentData, err := h.playlistService.GetSegmentFile(uint(videoId), resolution, segment)
	if err != nil {
		logrus.Errorf("Failed to get segment file: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	c.Header("Content-Type", "video/mp2t")
	c.Data(http.StatusOK, "video/mp2t", segmentData)
}

// Get transcode queue endpoint
func (h *Handlers) GetTranscodeQueue(c *gin.Context) {
	queue, err := h.transcodeService.GetTranscodeQueue()
	if err != nil {
		logrus.Errorf("Failed to get transcode queue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transcode queue"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// Get transcode status endpoint
func (h *Handlers) GetTranscodeStatus(c *gin.Context) {
	status, err := h.transcodeService.GetTranscodeStatus()
	if err != nil {
		logrus.Errorf("Failed to get transcode status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transcode status"})
		return
	}

	c.JSON(http.StatusOK, status)
}
