# 🚀 **PANDUAN MANUAL SETUP LINIER CHANNEL (TANPA DOCKER)**

## 📋 **Overview**
Panduan lengkap untuk menjalankan Linier Channel Video Transcoding Service secara manual tanpa Docker, menggunakan MySQL dan Redis lokal.

## 🔧 **Prerequisites**

### **1. Install Dependencies**
```bash
# Install Go, MySQL, Redis, FFmpeg
brew install go mysql redis ffmpeg

# Start services
brew services start mysql redis
```

### **2. Verify Installation**
```bash
# Check versions
go version
mysql --version
redis-server --version
ffmpeg -version
```

## 🗄️ **Database Setup**

### **1. Setup MySQL Password**
```bash
# Test MySQL connection (gunakan password yang sama dengan Navicat)
mysql -u root -p'Nd45mulh0!' -e "SELECT 1;"
```

### **2. Create Database**
```bash
# Create database
mysql -u root -p'Nd45mulh0!' -e "CREATE DATABASE IF NOT EXISTS linier_channel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### **3. Run Migrations**
```bash
# Run database migrations
mysql -u root -p'Nd45mulh0!' linier_channel < migrations/001_create_initial_schema.sql
```

## ⚙️ **Configuration**

### **1. Environment Variables**
Buat file `config.manual.env`:
```env
# Linier Channel Configuration - Manual Setup

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=Nd45mulh0!
DB_NAME=linier_channel
DB_CHARSET=utf8mb4

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Storage Configuration
UPLOAD_PATH=./uploads
TRANCODED_PATH=./transcoded-videos
ARCHIVE_PATH=./archive
MAX_FILE_SIZE=1073741824
ALLOWED_FORMATS=mp4,avi,mov,mkv

# FFmpeg Configuration (using system FFmpeg)
FFMPEG_PATH=/opt/homebrew/bin/ffmpeg
FFPROBE_PATH=/opt/homebrew/bin/ffprobe

# Transcoding Configuration
TRANSCODE_WORKERS=3
HLS_WINDOW=15
SEGMENT_TIME=4

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Kubernetes Configuration
POD_NAME=
NAMESPACE=default
```

### **2. Create Directories**
```bash
# Create required directories
mkdir -p uploads transcoded-videos archive
```

## 🔨 **Build & Run**

### **1. Build Application**
```bash
# Download dependencies
go mod tidy

# Build application
go build -o bin/linier-channel cmd/main.go
```

### **2. Run Application**
```bash
# Run with environment variables
DB_HOST=localhost \
DB_PORT=3306 \
DB_USER=root \
DB_PASSWORD='Nd45mulh0!' \
DB_NAME=linier_channel \
DB_CHARSET=utf8mb4 \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
REDIS_PASSWORD= \
REDIS_DB=0 \
UPLOAD_PATH=./uploads \
TRANCODED_PATH=./transcoded-videos \
ARCHIVE_PATH=./archive \
MAX_FILE_SIZE=1073741824 \
ALLOWED_FORMATS=mp4,avi,mov,mkv \
FFMPEG_PATH=/opt/homebrew/bin/ffmpeg \
FFPROBE_PATH=/opt/homebrew/bin/ffprobe \
TRANSCODE_WORKERS=3 \
HLS_WINDOW=15 \
SEGMENT_TIME=4 \
LOG_LEVEL=info \
LOG_FORMAT=json \
POD_NAME= \
NAMESPACE=default \
./bin/linier-channel > app.log 2>&1 &
```

### **3. Verify Application**
```bash
# Check if app is running
ps aux | grep linier-channel | grep -v grep

# Check port 8080
lsof -i :8080

# Test health endpoint
curl http://localhost:8080/health

# Test API endpoint
curl http://localhost:8080/api/v1/videos/
```

## 🧪 **Testing**

### **1. Test Video Upload**
```bash
# Copy video to uploads directory
cp your-video.mp4 uploads/

# FTP Watcher will automatically detect and process
```

### **2. Monitor Progress**
```bash
# Check application logs
tail -f app.log

# Check all videos
curl http://localhost:8080/api/v1/videos/

# Check specific video status
curl http://localhost:8080/api/v1/videos/1/status

# Check transcoding queue
curl http://localhost:8080/api/v1/admin/transcode/queue
```

### **3. Check Transcoding Results**
```bash
# Check transcoded files
ls -la transcoded-videos/

# Check HLS playlists
ls -la transcoded-videos/*/master.m3u8
ls -la transcoded-videos/*/720p/playlist.m3u8
```

## 📁 **Directory Structure**

```
/Users/herihandoko/Sites/transcoder/
├── uploads/                    # FTP Watcher monitors this
│   ├── your-video.mp4         # Upload videos here
│   └── test-video.avi
├── transcoded-videos/          # Transcoding output
│   ├── 2025/09/               # Date-based structure
│   │   └── video-name/
│   │       ├── master.m3u8    # Master playlist
│   │       ├── 720p/          # 720p profile
│   │       │   ├── playlist.m3u8
│   │       │   └── *.ts       # Video segments
│   │       ├── 480p/          # 480p profile
│   │       └── 360p/          # 360p profile
│   └── archive/               # Archived original files
├── bin/
│   └── linier-channel         # Application binary
├── app.log                    # Application logs
└── config.manual.env          # Environment variables
```

## 🔍 **API Endpoints**

### **Video Management**
```bash
# List all videos
GET /api/v1/videos/

# Get specific video
GET /api/v1/videos/{id}

# Get video status
GET /api/v1/videos/{id}/status

# Delete video
DELETE /api/v1/videos/{id}
```

### **Playlist Management**
```bash
# List playlists
GET /api/v1/playlists/

# Create playlist
POST /api/v1/playlists/

# Get playlist
GET /api/v1/playlists/{id}

# Update playlist
PUT /api/v1/playlists/{id}

# Delete playlist
DELETE /api/v1/playlists/{id}
```

### **Streaming**
```bash
# Get master playlist
GET /api/v1/stream/{videoId}/master.m3u8

# Get profile playlist
GET /api/v1/stream/{videoId}/{resolution}/playlist.m3u8

# Get video segment
GET /api/v1/stream/{videoId}/{resolution}/{segment}
```

### **Admin**
```bash
# Get transcoding queue
GET /api/v1/admin/transcode/queue

# Get transcoding status
GET /api/v1/admin/transcode/status
```

## 🚨 **Troubleshooting**

### **1. Transcoding Jobs Stuck in "uploaded" Status**
```bash
# Check if jobs are in database but not processing
curl -s http://localhost:8080/api/v1/admin/transcode/queue | jq '.[] | {id, status}'

# Check if jobs are in Redis queue
redis-cli llen "transcode_queue"

# If jobs are in database but not Redis, there's a queue issue
# Check application logs for Redis connection errors
tail -100 app.log | grep -i redis

# Manual fix: Restart application to re-queue jobs
pkill -f linier-channel
# Then restart application
```

### **2. Database Connection Issues**
```bash
# Check MySQL status
brew services list | grep mysql

# Test connection
mysql -u root -p'Nd45mulh0!' -e "SELECT 1;"

# Check database exists
mysql -u root -p'Nd45mulh0!' -e "SHOW DATABASES;" | grep linier_channel
```

### **2. Redis Connection Issues**
```bash
# Check Redis status
brew services list | grep redis

# Test Redis connection
redis-cli ping
```

### **3. Port Conflicts**
```bash
# Check what's using port 8080
lsof -i :8080

# Kill process using port 8080
lsof -ti :8080 | xargs kill -9
```

### **4. FFmpeg Issues**
```bash
# Check FFmpeg installation
ffmpeg -version
ffprobe -version

# Check paths in config
echo $FFMPEG_PATH
echo $FFPROBE_PATH
```

### **5. Application Issues**
```bash
# Check application logs
tail -f app.log

# Check if app is running
ps aux | grep linier-channel

# Restart application
pkill -f linier-channel
# Then run again with environment variables
```

## 🔄 **Daily Operations**

### **1. Start Application**
```bash
# Start services
brew services start mysql redis

# Run application
DB_HOST=localhost DB_PORT=3306 DB_USER=root DB_PASSWORD='Nd45mulh0!' DB_NAME=linier_channel DB_CHARSET=utf8mb4 REDIS_HOST=localhost REDIS_PORT=6379 REDIS_PASSWORD= REDIS_DB=0 UPLOAD_PATH=./uploads TRANCODED_PATH=./transcoded-videos ARCHIVE_PATH=./archive MAX_FILE_SIZE=1073741824 ALLOWED_FORMATS=mp4,avi,mov,mkv FFMPEG_PATH=/opt/homebrew/bin/ffmpeg FFPROBE_PATH=/opt/homebrew/bin/ffprobe TRANSCODE_WORKERS=3 HLS_WINDOW=15 SEGMENT_TIME=4 LOG_LEVEL=info LOG_FORMAT=json POD_NAME= NAMESPACE=default ./bin/linier-channel > app.log 2>&1 &
```

### **2. Stop Application**
```bash
# Stop application
pkill -f linier-channel

# Stop services (optional)
brew services stop mysql redis
```

### **3. Monitor Application**
```bash
# Check logs
tail -f app.log

# Check status
curl http://localhost:8080/health

# Check videos
curl http://localhost:8080/api/v1/videos/
```

## 📊 **Performance Monitoring**

### **1. System Resources**
```bash
# Check CPU usage
top -p $(pgrep linier-channel)

# Check memory usage
ps aux | grep linier-channel

# Check disk usage
du -sh uploads/ transcoded-videos/ archive/
```

### **2. Database Monitoring**
```bash
# Check database size
mysql -u root -p'Nd45mulh0!' -e "SELECT table_schema AS 'Database', ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)' FROM information_schema.tables WHERE table_schema = 'linier_channel' GROUP BY table_schema;"

# Check table sizes
mysql -u root -p'Nd45mulh0!' linier_channel -e "SELECT table_name, ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)' FROM information_schema.tables WHERE table_schema = 'linier_channel' ORDER BY (data_length + index_length) DESC;"
```

### **3. Redis Monitoring**
```bash
# Check Redis info
redis-cli info

# Check Redis memory usage
redis-cli info memory
```

## 🎯 **Production Considerations**

### **1. Security**
- Change default MySQL password
- Use environment variables for sensitive data
- Implement proper authentication
- Use HTTPS in production

### **2. Performance**
- Increase transcoding workers based on CPU cores
- Use SSD storage for better I/O performance
- Monitor disk space regularly
- Implement proper logging and monitoring

### **3. Backup**
- Regular database backups
- Backup transcoded videos
- Implement disaster recovery procedures

## 📝 **Notes**

- **FTP Watcher**: Automatically detects new files in `uploads/` directory
- **Date-based Structure**: Transcoding output organized by year/month
- **HLS Streaming**: Generates master and profile playlists
- **Archive**: Original files moved to archive after successful transcoding
- **Multi-format Support**: MP4, AVI, MOV, MKV, WMV, FLV, WebM
- **Auto-scaling**: Can be scaled horizontally in production

## 🆘 **Support**

Jika mengalami masalah:
1. Check logs: `tail -f app.log`
2. Verify services: `brew services list`
3. Test connections: MySQL, Redis, FFmpeg
4. Check disk space: `df -h`
5. Monitor resources: `top`, `htop`

---

**Manual setup berhasil!** 🎉 Sekarang Anda bisa upload video apapun ke direktori `uploads/` dan sistem akan otomatis memprosesnya.
