# Linier Channel - Video Transcoding Service

## 📋 Project Overview

**Linier Channel** adalah high-performance video transcoding service yang dibangun dengan Go, mendukung HLS streaming dan deployment Kubernetes. Service ini dirancang untuk menangani video transcoding secara otomatis dengan multiple quality profiles dan real-time progress tracking.

## 🚀 Key Features

- **Multi-Format Support**: MP4, AVI, MOV, MKV
- **Auto Transcoding**: 3 quality profiles (720p, 480p, 360p)
- **HLS Streaming**: Generate HLS playlists dan segments
- **Real-time Progress**: Live transcoding progress monitoring
- **Playlist Management**: Create dan manage video playlists
- **Kubernetes Ready**: Full Kubernetes deployment support
- **Horizontal Scaling**: Auto-scaling berdasarkan resource usage
- **Redis Queue**: Job queue management untuk transcoding tasks

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Upload API    │    │  Transcode      │    │   HLS Stream    │
│                 │───▶│   Workers       │───▶│                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     MySQL       │    │     Redis       │    │   File Storage  │
│   (Database)    │    │   (Queue)       │    │   (NFS/S3)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Tech Stack

- **Backend**: Go 1.21+
- **Database**: MySQL 8.0+
- **Cache/Queue**: Redis 7+
- **Video Processing**: FFmpeg
- **Container**: Docker + Docker Compose
- **Orchestration**: Kubernetes
- **Storage**: Persistent Volumes

## 📊 Transcoding Flow

1. **Upload**: User upload video via FTP/API
2. **Storage**: File disimpan ke directory/OSS
3. **Queue**: Service transcode membaca directory
4. **Processing**: Progress transcode (insert ke tabel videos)
5. **Output**: Generate 3 profile (720p, 480p, 360p) dengan file *.ts
6. **Update**: Update tabel video dengan status completed

## ⚙️ Configuration

### Transcode Settings
- **Video Codec**: H264
- **Audio Codec**: AAC (128k)
- **Segment Time**: 4 seconds
- **Profiles**: 720p (2000k), 480p (1000k), 360p (500k)

### Environment Variables
```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=mysql
DB_USER=root
DB_PASSWORD=password
DB_NAME=linier_channel

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Storage
UPLOAD_PATH=/uploads
TRANCODED_PATH=/transcoded-videos
MAX_FILE_SIZE=1073741824

# Transcoding
TRANSCODE_WORKERS=3
SEGMENT_TIME=4
```

## 🐳 Docker Deployment

### Services Structure
- **Service Transcode**: Custom image (Ubuntu 22.04 + FFmpeg)
- **Redis**: redis:7-alpine
- **MySQL**: mysql:8.0 (existing image)

### Quick Start
```bash
# Clone repository
git clone <repository-url>
cd transcoder

# Build and run
docker-compose up -d

# Check status
docker-compose ps
```

## ☸️ Kubernetes Deployment

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n linier-channel

# Access service
kubectl port-forward svc/linier-channel-service 8080:8080 -n linier-channel
```

## 📡 API Endpoints

### Video Management
- `POST /api/v1/videos/upload` - Upload video
- `GET /api/v1/videos/{id}/status` - Get video status
- `GET /api/v1/videos/{id}/profiles` - Get transcoding profiles

### Streaming
- `GET /api/v1/stream/{videoId}/master.m3u8` - Master playlist
- `GET /api/v1/stream/{videoId}/{profile}/playlist.m3u8` - Profile playlist
- `GET /api/v1/stream/{videoId}/{profile}/{segment}.ts` - Video segments

### Playlist Management
- `POST /api/v1/playlists` - Create playlist
- `GET /api/v1/playlists` - List playlists
- `POST /api/v1/playlists/{id}/videos` - Add video to playlist

## 📁 Project Structure

```
transcoder/
├── cmd/main.go                 # Application entry point
├── internal/
│   ├── config/                # Configuration management
│   ├── database/              # Database connection
│   ├── handlers/              # HTTP handlers
│   ├── models/                # Data models
│   ├── services/              # Business logic
│   └── worker/                # Background workers
├── k8s/                       # Kubernetes manifests
├── migrations/                # Database migrations
├── Dockerfile                 # Docker configuration
├── docker-compose.yml         # Docker Compose setup
└── README.md                  # Documentation
```

## 🔍 Monitoring & Health Checks

- **Health Endpoint**: `/health`
- **Transcode Status**: `/api/v1/admin/transcode/status`
- **Queue Status**: `/api/v1/admin/queue/status`
- **Worker Status**: `/api/v1/admin/workers/status`

## 🚀 Production Considerations

1. **Storage**: Use persistent volumes untuk video storage
2. **Scaling**: Configure HPA untuk automatic scaling
3. **Monitoring**: Set up monitoring dan alerting
4. **Security**: Use secrets untuk sensitive configuration
5. **Backup**: Regular database dan storage backups

## 📝 Development

### Prerequisites
- Go 1.21+
- MySQL 8.0+
- Redis 7+
- FFmpeg

### Build & Run
```bash
# Install dependencies
go mod download

# Build
go build -o bin/linier-channel cmd/main.go

# Run
./bin/linier-channel
```

### Testing
```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## 📞 Support

For support and questions:
- Create issue on Bitbucket
- Check documentation
- Review troubleshooting guide

---

**Built with ❤️ for high-performance video transcoding**
