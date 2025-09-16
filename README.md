# Linier Channel - Video Transcoding Service

A high-performance video transcoding service built with Go, supporting Kubernetes deployment and HLS streaming.

## Features

- **Video Upload**: Support for MP4, AVI, MOV, MKV formats
- **Auto Transcoding**: Automatic transcoding to multiple quality profiles (720p, 480p, 360p)
- **HLS Streaming**: Generate HLS playlists and segments for streaming
- **Progress Tracking**: Real-time transcoding progress monitoring
- **Playlist Management**: Create and manage video playlists
- **Kubernetes Ready**: Full Kubernetes deployment support
- **Horizontal Scaling**: Auto-scaling based on resource usage
- **Redis Queue**: Job queue management for transcoding tasks

## Architecture

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

## Quick Start

### Development Setup (with Docker Compose)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd transcoder
   ```

2. **Start all services (MySQL, Redis, App)**
   ```bash
   docker-compose up -d
   ```

3. **Check service status**
   ```bash
   docker-compose ps
   ```

4. **Access the API**
   ```bash
   curl http://localhost:8080/health
   ```

### Production Setup (External Database)

1. **Configure environment**
   ```bash
   cp config.production.env .env
   # Edit .env with your database credentials
   ```

2. **Deploy application only**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **Check deployment**
   ```bash
   docker-compose -f docker-compose.prod.yml ps
   ```

> **Note**: For production deployment with external MySQL and Redis, see [DEPLOYMENT.md](DEPLOYMENT.md) for detailed instructions.

### Using Kubernetes

1. **Deploy to Kubernetes**
   ```bash
   kubectl apply -f k8s/
   ```

2. **Check deployment status**
   ```bash
   kubectl get pods -n linier-channel
   ```

3. **Access the service**
   ```bash
   kubectl port-forward svc/linier-channel-service 8080:8080 -n linier-channel
   ```

## API Documentation

### Upload Video

```bash
curl -X POST http://localhost:8080/api/v1/videos/upload \
  -F "file=@video.mp4"
```

### Get Video Status

```bash
curl http://localhost:8080/api/v1/videos/1/status
```

### Get HLS Playlist

```bash
curl http://localhost:8080/api/v1/stream/1/master.m3u8
```

### Create Playlist

```bash
curl -X POST http://localhost:8080/api/v1/playlists \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Playlist",
    "description": "A collection of videos",
    "video_ids": [1, 2, 3]
  }'
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | 8080 |
| `DB_HOST` | Database host | localhost |
| `DB_PASSWORD` | Database password | password |
| `REDIS_HOST` | Redis host | localhost |
| `UPLOAD_PATH` | Upload directory | /uploads |
| `TRANCODED_PATH` | Transcoded videos directory | /transcoded-videos |
| `MAX_FILE_SIZE` | Maximum file size (bytes) | 1073741824 |
| `TRANSCODE_WORKERS` | Number of transcode workers | 3 |

### Database Schema

The service uses MySQL with the following tables:
- `videos` - Master video table
- `video_profiles` - Video profile configurations
- `playlists` - Playlist management
- `playlist_videos` - Playlist-video relationships
- `transcode_jobs` - Job queue management
- `system_config` - System configuration

## Development

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 7+
- FFmpeg

### Build

```bash
go mod download
go build -o bin/linier-channel cmd/main.go
```

### Run

```bash
./bin/linier-channel
```

### Test

```bash
go test ./...
```

## Deployment

### Docker

```bash
docker build -t linier-channel .
docker run -p 8080:8080 linier-channel
```

### Kubernetes

```bash
kubectl apply -f k8s/
```

### Production Considerations

1. **Storage**: Use persistent volumes for video storage
2. **Scaling**: Configure HPA for automatic scaling
3. **Monitoring**: Set up monitoring and alerting
4. **Security**: Use secrets for sensitive configuration
5. **Backup**: Regular database and storage backups

## Monitoring

### Health Checks

- `/health` - Application health status
- `/api/v1/admin/transcode/status` - Transcoding status

### Logs

```bash
# Docker Compose
docker-compose logs -f app

# Kubernetes
kubectl logs -f deployment/linier-channel -n linier-channel
```

## Troubleshooting

### Common Issues

1. **FFmpeg not found**
   - FFmpeg is already installed in the Docker container
   - Check FFMPEG_PATH environment variable (should be `/usr/bin/ffmpeg`)
   - Verify container is running: `docker exec -it linier-channel-app ffmpeg -version`

2. **Database connection failed**
   - Verify database credentials
   - Check database service status

3. **Redis connection failed**
   - Verify Redis service status
   - Check Redis configuration

4. **Storage issues**
   - Check volume mounts
   - Verify storage permissions

### Debug Mode

Set `LOG_LEVEL=debug` for detailed logging.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the troubleshooting guide
