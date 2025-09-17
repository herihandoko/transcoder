package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Storage    StorageConfig
	FFmpeg     FFmpegConfig
	Transcode  TranscodeConfig
	Logging    LoggingConfig
	Kubernetes KubernetesConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Charset  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type StorageConfig struct {
	UploadPath     string
	TranscodedPath string
	ArchivePath    string
	MaxFileSize    int64
	AllowedFormats []string
}

type FFmpegConfig struct {
	FFmpegPath  string
	FFprobePath string
}

type TranscodeConfig struct {
	Workers     int
	HLSWindow   int
	SegmentTime int
}

type LoggingConfig struct {
	Level  string
	Format string
}

type KubernetesConfig struct {
	PodName   string
	Namespace string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "Nd45mulh0!"),
			Name:     getEnv("DB_NAME", "linier_channel"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Storage: StorageConfig{
			UploadPath:     getEnv("UPLOAD_PATH", "/uploads"),
			TranscodedPath: getEnv("TRANCODED_PATH", "/transcoded-videos"),
			ArchivePath:    getEnv("ARCHIVE_PATH", "/archive"),
			MaxFileSize:    getEnvAsInt64("MAX_FILE_SIZE", 1073741824), // 1GB
			AllowedFormats: strings.Split(getEnv("ALLOWED_FORMATS", "mp4,avi,mov,mkv"), ","),
		},
		FFmpeg: FFmpegConfig{
			FFmpegPath:  getEnv("FFMPEG_PATH", "/usr/bin/ffmpeg"),
			FFprobePath: getEnv("FFPROBE_PATH", "/usr/bin/ffprobe"),
		},
		Transcode: TranscodeConfig{
			Workers:     getEnvAsInt("TRANSCODE_WORKERS", 3),
			HLSWindow:   getEnvAsInt("HLS_WINDOW", 15),
			SegmentTime: getEnvAsInt("SEGMENT_TIME", 4),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Kubernetes: KubernetesConfig{
			PodName:   getEnv("POD_NAME", ""),
			Namespace: getEnv("NAMESPACE", "default"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
