# Docker Local Setup Guide

## Overview
This setup runs the Linier Channel application in Docker while using local MySQL and Redis services.

## Prerequisites

### 1. Local Services Required
- **MySQL** running on `localhost:3306`
  - Database: `linier_channel`
  - User: `root`
  - Password: `Nd45mulh0!`
- **Redis** running on `localhost:6379`

### 2. Docker Requirements
- Docker
- Docker Compose

## Quick Start

### 1. Start Local Services
Make sure MySQL and Redis are running locally:

```bash
# Check MySQL
mysql -u root -p'Nd45mulh0!' -e "SHOW DATABASES;"

# Check Redis
redis-cli ping
```

### 2. Run with Script
```bash
./run-docker-local.sh
```

### 3. Manual Docker Commands
```bash
# Build and start
docker-compose -f docker-compose.local.yml up --build -d

# View logs
docker-compose -f docker-compose.local.yml logs -f app

# Stop
docker-compose -f docker-compose.local.yml down
```

## Configuration

### Environment Variables
The application uses these key configurations:

- **Database**: `host.docker.internal:3306` (local MySQL)
- **Redis**: `host.docker.internal:6379` (local Redis)
- **Upload Path**: `/Users/herihandoko/Documents/Transcode/input` (mounted to container)
- **Transcoded Path**: `/Users/herihandoko/Documents/Transcode/output` (mounted to container)
- **Archive Path**: `/Users/herihandoko/Documents/Transcode/archive` (mounted to container)

### Directory Structure
```
/Users/herihandoko/Documents/Transcode/
├── input/                # Uploaded videos
├── output/               # HLS output
└── archive/              # Archived originals

/Users/herihandoko/Sites/transcoder/
└── docker-compose.local.yml
```

## Access Points

- **Application**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **API**: http://localhost:8080/api/v1
- **Nginx (if enabled)**: http://localhost:80

## Troubleshooting

### MySQL Connection Issues
```bash
# Check if MySQL is accessible from Docker
docker run --rm --network host mysql:8.0 mysql -h localhost -u root -p'Nd45mulh0!' -e "SHOW DATABASES;"
```

### Redis Connection Issues
```bash
# Check if Redis is accessible from Docker
docker run --rm --network host redis:7-alpine redis-cli -h localhost ping
```

### Permission Issues
```bash
# Fix directory permissions
sudo chown -R $USER:$USER uploads transcoded-videos archive
chmod 755 uploads transcoded-videos archive
```

### View Logs
```bash
# Application logs
docker-compose -f docker-compose.local.yml logs -f app

# All services logs
docker-compose -f docker-compose.local.yml logs -f
```

## File Structure

### docker-compose.local.yml
- Runs only the app and nginx containers
- Uses `host.docker.internal` to access local MySQL/Redis
- Mounts local directories for file storage

### nginx.conf
- Proxies API requests to app container
- Serves transcoded videos directly
- Includes CORS headers for video streaming

### run-docker-local.sh
- Automated setup script
- Checks local services
- Creates directories
- Starts containers

## Notes

- The application will create the database schema automatically on first run
- FFmpeg is included in the Docker image
- All file operations use the mounted local directories
- The setup preserves your existing MySQL data and Redis cache
