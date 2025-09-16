# FTP Watcher - Linier Channel

## 📋 Overview

FTP Watcher adalah komponen yang memantau direktori `uploads/` untuk file video baru yang di-upload via FTP atau metode lain. Ketika file video terdeteksi, sistem akan otomatis memproses file tersebut untuk transcoding.

## 🚀 Features

- **Auto Detection**: Mendeteksi file video baru secara otomatis
- **Multi Format Support**: MP4, AVI, MOV, MKV, WMV, FLV, WebM
- **Auto Processing**: Otomatis membuat record database dan queue transcoding
- **File Validation**: Validasi file sebelum processing
- **Graceful Shutdown**: Support graceful shutdown

## 🔧 How It Works

### 1. File Monitoring
```
uploads/ (monitored directory)
├── video1.mp4 ← FTP Watcher detects
├── video2.avi ← FTP Watcher detects
└── document.pdf ← Ignored (not video)
```

### 2. Processing Flow
```
File Detected → Validate → Create DB Record → Queue Transcoding → Archive Original
```

### 3. Supported File Types
- `.mp4` - MP4 Video
- `.avi` - AVI Video  
- `.mov` - QuickTime Movie
- `.mkv` - Matroska Video
- `.wmv` - Windows Media Video
- `.flv` - Flash Video
- `.webm` - WebM Video

## ⚙️ Configuration

### Environment Variables
```bash
# Storage Configuration
UPLOAD_PATH=/uploads                    # Directory yang dimonitor
TRANCODED_PATH=/transcoded-videos      # Output transcoding
ARCHIVE_PATH=/archive                  # Archive file asli

# Database & Redis (untuk processing)
DB_HOST=localhost
DB_PORT=3306
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Directory Structure
```
uploads/                    # FTP Watcher monitors this
├── video1.mp4            # Auto detected & processed
├── video2.avi            # Auto detected & processed
└── temp/                 # Temporary files (ignored)

transcoded-videos/         # Output after transcoding
├── 2025/09/video1/
│   ├── 720p/
│   ├── 480p/
│   └── 360p/
└── 2025/09/video2/
    ├── 720p/
    ├── 480p/
    └── 360p/

archive/                   # Original files after transcoding
├── 2025/09/video1/
│   └── video1.mp4
└── 2025/09/video2/
    └── video2.avi
```

## 🚀 Usage

### 1. Start Application
```bash
# FTP Watcher akan berjalan otomatis
./bin/linier-channel
```

### 2. Upload Video via FTP
```bash
# Setup FTP server (contoh: vsftpd)
# File yang di-upload akan masuk ke uploads/
# FTP Watcher akan otomatis mendeteksi dan memproses
```

### 3. Manual Upload Test
```bash
# Copy file video ke direktori uploads
cp /path/to/your/video.mp4 uploads/

# FTP Watcher akan otomatis mendeteksi dan memproses
```

## 📊 Monitoring

### 1. Check Logs
```bash
# Aplikasi logs
tail -f app.log

# FTP Watcher specific logs
grep "FTP Watcher" app.log
```

### 2. Check Status
```bash
# Check apakah aplikasi berjalan
ps aux | grep linier-channel

# Check direktori uploads
ls -la uploads/

# Check database untuk video baru
curl http://localhost:8080/api/v1/videos/
```

### 3. Check Transcoding Status
```bash
# Check transcoding queue
curl http://localhost:8080/api/v1/admin/transcode/queue

# Check transcoding status
curl http://localhost:8080/api/v1/admin/transcode/status
```

## 🔍 Troubleshooting

### 1. FTP Watcher Tidak Berjalan
```bash
# Check log aplikasi
tail -f app.log

# Check apakah direktori uploads ada
ls -la uploads/

# Check permission direktori
chmod 755 uploads/
```

### 2. File Tidak Terdeteksi
```bash
# Pastikan file extension didukung
# (mp4, avi, mov, mkv, wmv, flv, webm)

# Check apakah file sudah fully written
# (FTP Watcher menunggu 2 detik sebelum processing)

# Check file size (tidak boleh 0 bytes)
ls -la uploads/
```

### 3. Database Error
```bash
# Check database connection
# Check apakah tabel videos dan video_profiles ada
# Check Redis connection
```

### 4. Transcoding Error
```bash
# Check FFmpeg installation
ffmpeg -version

# Check disk space
df -h

# Check file permissions
ls -la uploads/
```

## 🛠️ Development

### 1. Test FTP Watcher
```bash
# Build aplikasi
go build -o bin/linier-channel cmd/main.go

# Run aplikasi
./bin/linier-channel

# Test dengan upload file
cp test-video.mp4 uploads/

# Check logs
tail -f app.log
```

### 2. Customize File Types
Edit `internal/services/ftp_watcher.go`:
```go
func (fw *FTPWatcher) isVideoFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    allowedExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}
    // Add more extensions here
    return contains(allowedExts, ext)
}
```

### 3. Customize Processing
Edit `internal/services/ftp_watcher.go`:
```go
func (fw *FTPWatcher) processNewVideo(filePath string) {
    // Add custom processing logic here
    // e.g., file validation, metadata extraction, etc.
}
```

## 📝 API Integration

FTP Watcher terintegrasi dengan API endpoints:

### Video Management
- `GET /api/v1/videos/` - List semua videos (termasuk yang dari FTP)
- `GET /api/v1/videos/{id}/status` - Status transcoding
- `GET /api/v1/videos/{id}/profiles` - Transcoding profiles

### Streaming
- `GET /api/v1/stream/{videoId}/master.m3u8` - Master playlist
- `GET /api/v1/stream/{videoId}/{profile}/playlist.m3u8` - Profile playlist
- `GET /api/v1/stream/{videoId}/{profile}/{segment}.ts` - Video segments

## 🔒 Security Considerations

1. **File Validation**: Validasi file sebelum processing
2. **Path Security**: Sanitize file paths
3. **Permission**: Proper file permissions
4. **Size Limits**: File size validation
5. **Type Validation**: Only process video files

## 📈 Performance

- **Concurrent Processing**: Multiple files dapat diproses bersamaan
- **Queue Management**: Redis queue untuk transcoding jobs
- **Resource Management**: Proper resource cleanup
- **Error Handling**: Robust error handling dan recovery

## 🚀 Production Deployment

### 1. Docker
```bash
# Build image
docker build -t linier-channel .

# Run container
docker run -d -p 8080:8080 linier-channel
```

### 2. Kubernetes
```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n linier-channel
```

### 3. Process Manager
```bash
# Using systemd
sudo systemctl start linier-channel
sudo systemctl enable linier-channel

# Using PM2
pm2 start ./bin/linier-channel --name linier-channel
```

---

**FTP Watcher sudah terintegrasi dan akan berjalan otomatis ketika aplikasi dijalankan!**
