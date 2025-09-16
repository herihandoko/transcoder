# Database Migrations

This directory contains database migration scripts for the Linier Channel video transcoding service.

## Migration Files

### 001_create_initial_schema.sql
Creates the initial database schema with all required tables:
- `videos` - Master video table
- `video_profiles` - Video profile configurations and status
- `playlists` - Playlist management
- `playlist_videos` - Playlist-video relationships
- `transcode_jobs` - Job queue management
- `system_config` - System configuration

### 002_add_sample_data.sql
Adds sample data for testing and development:
- Sample playlists
- Sample videos
- Sample video profiles
- Sample transcode jobs

## Running Migrations

### Using MySQL Command Line
```bash
# Run all migrations
mysql -u username -p < migrations/001_create_initial_schema.sql
mysql -u username -p < migrations/002_add_sample_data.sql
```

### Using Docker
```bash
# Run migrations in Docker container
docker exec -i mysql_container mysql -u username -p < migrations/001_create_initial_schema.sql
```

### Using Application
```go
// Run migrations programmatically
func RunMigrations() error {
    // Read and execute migration files
    // Handle versioning and rollbacks
}
```

## Database Schema Overview

```
videos (1) -----> (N) video_profiles
  |                    |
  |                    v
  |              transcode_jobs
  |
  v
playlist_videos (N) <-- (1) playlists
```

## Configuration

Default system configuration is inserted in migration 001:
- `transcode_workers`: Number of concurrent workers
- `max_file_size`: Maximum file size limit
- `allowed_formats`: Supported video formats
- `storage_path`: Path for transcoded videos
- `hls_window`: HLS window size
- `ffmpeg_path`: FFmpeg binary location
- `redis_url`: Redis connection URL
- `mysql_url`: MySQL connection URL

## Indexes

The schema includes optimized indexes for:
- Status-based queries
- Date-based queries
- Foreign key relationships
- Resolution-based filtering
- Priority-based job processing

## Foreign Key Constraints

All foreign keys use `ON DELETE CASCADE` to maintain data integrity when records are deleted.
