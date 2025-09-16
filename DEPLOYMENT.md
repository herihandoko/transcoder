# Linier Channel - Production Deployment Guide

## Overview
This guide covers deploying Linier Channel in production with external MySQL and Redis databases.

## Prerequisites
- Docker and Docker Compose installed
- External MySQL database (8.0+)
- External Redis server (6.0+)
- FFmpeg installed (see FFmpeg Installation section below)

## FFmpeg Installation

### Option 1: Install FFmpeg in Docker Container (Recommended)

Update your `Dockerfile` to include FFmpeg:

```dockerfile
FROM golang:1.21-alpine AS builder

# Install FFmpeg and dependencies
RUN apk add --no-cache ffmpeg

# ... rest of your Dockerfile
```

### Option 2: Install FFmpeg on Host System

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**CentOS/RHEL:**
```bash
sudo yum install epel-release
sudo yum install ffmpeg
```

**macOS:**
```bash
brew install ffmpeg
```

**Windows:**
Download from https://ffmpeg.org/download.html

### Option 3: Use FFmpeg Docker Image

Mount FFmpeg from a separate container:

```yaml
# docker-compose.prod.yml
services:
  app:
    # ... your app config
    volumes:
      - /usr/bin/ffmpeg:/usr/bin/ffmpeg:ro
      - /usr/bin/ffprobe:/usr/bin/ffprobe:ro
```

## Production Deployment

### 1. Environment Configuration

Copy the production environment template:
```bash
cp config.production.env .env
```

Update `.env` with your actual database and Redis credentials:
```bash
# Database Configuration
DB_HOST=your-mysql-host.com
DB_PORT=3306
DB_USER=your-mysql-user
DB_PASSWORD=your-mysql-password
DB_NAME=linier_channel

# Redis Configuration
REDIS_HOST=your-redis-host.com
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
```

### 2. Database Setup

Run the database migrations on your external MySQL:
```bash
# Connect to your MySQL database
mysql -h your-mysql-host.com -u your-mysql-user -p

# Create database (if not exists)
CREATE DATABASE IF NOT EXISTS linier_channel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# Run migrations
source migrations/001_create_initial_schema.sql
source migrations/002_add_sample_data.sql
```

### 3. Deploy Application

Use the production Docker Compose file:
```bash
# Deploy with production configuration
docker-compose -f docker-compose.prod.yml up -d

# Check logs
docker-compose -f docker-compose.prod.yml logs -f app
```

### 4. Health Check

Verify the application is running:
```bash
curl http://localhost:8080/health
```

### 5. Storage Volumes

The application uses the following volumes:
- `uploads_data`: For uploaded video files
- `transcoded_data`: For transcoded HLS files
- `archive_data`: For archived original files

### 6. Nginx Configuration (Optional)

If using Nginx for file serving, create `nginx.conf`:
```nginx
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    server {
        listen 80;
        
        location /transcoded/ {
            alias /usr/share/nginx/html/transcoded/;
            add_header Access-Control-Allow-Origin *;
            add_header Cache-Control "public, max-age=3600";
        }
        
        location / {
            proxy_pass http://app:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

## Kubernetes Deployment

For Kubernetes deployment, use the manifests in the `k8s/` directory:

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n linier-channel
kubectl get services -n linier-channel
```

## Monitoring

### Health Endpoints
- `GET /health` - Application health check
- `GET /api/v1/admin/transcode/status` - Transcoding status
- `GET /api/v1/admin/transcode/queue` - Job queue status

### Logs
```bash
# Application logs
docker-compose -f docker-compose.prod.yml logs -f app

# All services logs
docker-compose -f docker-compose.prod.yml logs -f
```

## Scaling

### Horizontal Scaling
- Use multiple application instances behind a load balancer
- Ensure shared storage for video files
- Use external Redis for job queue sharing

### Vertical Scaling
- Increase `TRANSCODE_WORKERS` for more concurrent transcoding
- Allocate more CPU/memory for FFmpeg processes

## Security Considerations

1. **Database Security**:
   - Use strong passwords
   - Enable SSL connections
   - Restrict database access to application servers only

2. **Redis Security**:
   - Use authentication
   - Enable TLS if possible
   - Restrict network access

3. **File Storage**:
   - Secure file uploads
   - Validate file types and sizes
   - Use proper file permissions

4. **Network Security**:
   - Use HTTPS in production
   - Implement rate limiting
   - Use proper firewall rules

## Troubleshooting

### Common Issues

1. **Database Connection Failed**:
   - Check database credentials
   - Verify network connectivity
   - Ensure database is running

2. **Redis Connection Failed**:
   - Check Redis credentials
   - Verify Redis is running
   - Check network connectivity

3. **Transcoding Fails**:
   - Verify FFmpeg is installed in Docker container: `docker exec -it linier-channel-app ffmpeg -version`
   - Check file permissions
   - Monitor disk space
   - Check FFmpeg paths in environment variables

4. **File Upload Issues**:
   - Check storage permissions
   - Verify disk space
   - Check file size limits

### Logs Location
- Application logs: Docker container logs
- Database logs: External MySQL logs
- Redis logs: External Redis logs

## Backup Strategy

1. **Database Backup**:
   ```bash
   mysqldump -h your-mysql-host.com -u your-mysql-user -p linier_channel > backup.sql
   ```

2. **File Storage Backup**:
   - Backup `uploads_data` volume
   - Backup `transcoded_data` volume
   - Backup `archive_data` volume

3. **Redis Backup**:
   - Use Redis persistence (RDB/AOF)
   - Regular backup of Redis data

## Performance Optimization

1. **Database**:
   - Use connection pooling
   - Optimize queries
   - Use proper indexes

2. **Redis**:
   - Use Redis clustering for high availability
   - Monitor memory usage
   - Use appropriate eviction policies

3. **Storage**:
   - Use SSD storage for better I/O
   - Implement CDN for video delivery
   - Use appropriate file system

4. **Transcoding**:
   - Use GPU acceleration if available
   - Optimize FFmpeg parameters
   - Use appropriate worker counts
