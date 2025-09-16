# Linier Channel - Video Transcoding Flow

## System Architecture Flow

```mermaid
graph TB
    A[Video Upload] --> B[Upload Service]
    B --> C[File Storage]
    C --> D[File Watcher]
    D --> E[Queue Job]
    E --> F[Transcode Service]
    F --> G[FFmpeg Processing]
    G --> H[Generate 720p Profile]
    G --> I[Generate 480p Profile]
    G --> J[Generate 360p Profile]
    H --> K[Create .ts Segments]
    I --> K
    J --> K
    K --> L[Generate .m3u8 Playlists]
    L --> M[Update Database]
    M --> N[Master Video Table]
    N --> O[Playlist Generator]
    O --> P[HLS Streaming Service]
    
    subgraph "Kubernetes Cluster"
        B
        D
        F
        P
    end
    
    subgraph "Storage"
        C
        K
        L
    end
    
    subgraph "Database"
        M
        N
    end
```

## Detailed Transcoding Process

```mermaid
sequenceDiagram
    participant U as User
    participant US as Upload Service
    participant FW as File Watcher
    participant Q as Queue
    participant TS as Transcode Service
    participant FF as FFmpeg
    participant DB as Database
    participant PS as Playlist Service
    
    U->>US: Upload MP4 Video
    US->>DB: Insert video record (status: uploaded)
    US->>FW: Notify new file
    FW->>Q: Add transcode job
    Q->>TS: Process job
    TS->>DB: Update status (processing)
    
    loop For each profile (720p, 480p, 360p)
        TS->>FF: Transcode to profile
        FF->>TS: Generate .ts segments
        TS->>DB: Update progress
    end
    
    TS->>PS: Generate playlists
    PS->>DB: Update status (completed)
    TS->>DB: Final status update
```

## Database Schema Relationship

```mermaid
erDiagram
    VIDEOS ||--o{ TRANCODE_PROGRESS : has
    VIDEOS ||--o{ PLAYLISTS : contains
    
    VIDEOS {
        int id PK
        string filename
        string original_path
        bigint file_size
        int duration
        enum status
        timestamp created_at
        timestamp updated_at
    }
    
    TRANCODE_PROGRESS {
        int id PK
        int video_id FK
        enum profile
        enum status
        int progress_percentage
        string output_path
        text error_message
        timestamp started_at
        timestamp completed_at
    }
    
    PLAYLISTS {
        int id PK
        string name
        text description
        json video_ids
        timestamp created_at
    }
```

## Kubernetes Deployment Architecture

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Upload Namespace"
            US[Upload Service Pod]
            FW[File Watcher Pod]
        end
        
        subgraph "Transcode Namespace"
            TS1[Transcode Worker 1]
            TS2[Transcode Worker 2]
            TS3[Transcode Worker N]
            Q[Redis Queue]
        end
        
        subgraph "Streaming Namespace"
            PS[Playlist Service]
            SS[Streaming Service]
        end
        
        subgraph "Storage"
            PV[Persistent Volume]
            CM[ConfigMap]
            SEC[Secret]
        end
        
        subgraph "Database"
            MYSQL[MySQL Pod]
        end
    end
    
    EXT[External Users] --> US
    US --> PV
    FW --> Q
    Q --> TS1
    Q --> TS2
    Q --> TS3
    TS1 --> PV
    TS2 --> PV
    TS3 --> PV
    TS1 --> MYSQL
    TS2 --> MYSQL
    TS3 --> MYSQL
    PS --> MYSQL
    SS --> PV
    EXT --> SS
```

## HLS Output Structure

```
/transcoded-videos/
├── video1/
│   ├── 720p/
│   │   ├── playlist.m3u8
│   │   ├── 00001.ts
│   │   ├── 00002.ts
│   │   └── ...
│   ├── 480p/
│   │   ├── playlist.m3u8
│   │   ├── 00001.ts
│   │   └── ...
│   └── 360p/
│       ├── playlist.m3u8
│       ├── 00001.ts
│       └── ...
└── master.m3u8
```

## URL Patterns

- Master Playlist: `{base_url}/videos/{video_id}/master.m3u8`
- Profile Playlist: `{base_url}/videos/{video_id}/{profile}/playlist.m3u8`
- Video Segments: `{base_url}/videos/{video_id}/{profile}/{segment}.ts`


