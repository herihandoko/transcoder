package worker

import (
	"context"
	"fmt"
	"linier-channel/internal/config"
	"linier-channel/internal/services"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type WorkerManager struct {
	transcodeService *services.TranscodeService
	redis            *redis.Client
	config           *config.Config
	workers          []*Worker
	stopChan         chan bool
}

type Worker struct {
	ID              int
	transcodeService *services.TranscodeService
	redis           *redis.Client
	stopChan        chan bool
}

func NewWorkerManager(transcodeService *services.TranscodeService, redis *redis.Client, cfg *config.Config) *WorkerManager {
	return &WorkerManager{
		transcodeService: transcodeService,
		redis:            redis,
		config:           cfg,
		stopChan:         make(chan bool),
	}
}

func (wm *WorkerManager) Start() {
	logrus.Info("Starting transcode workers...")

	// Create workers
	for i := 0; i < wm.config.Transcode.Workers; i++ {
		worker := &Worker{
			ID:              i + 1,
			transcodeService: wm.transcodeService,
			redis:           wm.redis,
			stopChan:        make(chan bool),
		}
		wm.workers = append(wm.workers, worker)

		// Start worker in goroutine
		go worker.Start()
	}

	// Wait for stop signal
	<-wm.stopChan
	logrus.Info("Stopping transcode workers...")

	// Stop all workers
	for _, worker := range wm.workers {
		worker.Stop()
	}
}

func (wm *WorkerManager) Stop() {
	close(wm.stopChan)
}

func (w *Worker) Start() {
	logrus.Infof("Worker %d started", w.ID)

	for {
		select {
		case <-w.stopChan:
			logrus.Infof("Worker %d stopped", w.ID)
			return
		default:
			// Try to get a job from the queue
			job, err := w.getNextJob()
			if err != nil {
				logrus.Errorf("Worker %d failed to get next job: %v", w.ID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No job available, sleep and try again
				time.Sleep(1 * time.Second)
				continue
			}

			// Process the job
			w.processJob(job)
		}
	}
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) getNextJob() (*JobData, error) {
	ctx := context.Background()
	
	// Try to get a job from Redis queue
	result, err := w.redis.BRPop(ctx, 5*time.Second, "transcode_queue").Result()
	if err != nil {
		if err == redis.Nil {
			// No job available
			return nil, nil
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid job data")
	}

	// Parse job data
	jobData := result[1]
	parts := strings.Split(jobData, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid job format")
	}

	videoID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid video ID: %v", err)
	}

	profileID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %v", err)
	}

	return &JobData{
		VideoID:   uint(videoID),
		ProfileID: uint(profileID),
	}, nil
}

func (w *Worker) processJob(job *JobData) {
	logrus.Infof("Worker %d processing job: video_id=%d, profile_id=%d", w.ID, job.VideoID, job.ProfileID)

	// Process the transcoding job
	err := w.transcodeService.ProcessTranscodeJob(job.VideoID, job.ProfileID)
	if err != nil {
		logrus.Errorf("Worker %d failed to process job: %v", w.ID, err)
		return
	}

	logrus.Infof("Worker %d completed job: video_id=%d, profile_id=%d", w.ID, job.VideoID, job.ProfileID)
}

type JobData struct {
	VideoID   uint `json:"video_id"`
	ProfileID uint `json:"profile_id"`
}

// QueueJob adds a job to the queue
func (wm *WorkerManager) QueueJob(videoID, profileID uint) error {
	ctx := context.Background()
	jobData := fmt.Sprintf("%d:%d", videoID, profileID)
	
	return wm.redis.LPush(ctx, "transcode_queue", jobData).Err()
}

// GetQueueStatus returns the current queue status
func (wm *WorkerManager) GetQueueStatus() (map[string]interface{}, error) {
	ctx := context.Background()
	
	// Get queue length
	queueLength, err := wm.redis.LLen(ctx, "transcode_queue").Result()
	if err != nil {
		return nil, err
	}

	// Get active workers count
	activeWorkers := len(wm.workers)

	return map[string]interface{}{
		"queue_length":    queueLength,
		"active_workers":  activeWorkers,
		"max_workers":     wm.config.Transcode.Workers,
	}, nil
}

// GetWorkerStatus returns the status of all workers
func (wm *WorkerManager) GetWorkerStatus() []WorkerStatus {
	var statuses []WorkerStatus
	
	for _, worker := range wm.workers {
		statuses = append(statuses, WorkerStatus{
			ID:     worker.ID,
			Status: "running",
		})
	}
	
	return statuses
}

type WorkerStatus struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}
