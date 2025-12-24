Directory structure plan:

```
bucket/
└── [videoId]/
    ├── source/
    │   └── original.mp4       <-- The original uploaded file
    │
    ├── hls/                     <-- All HLS files
    │   ├── master.m3u8
    │   ├── 1080p/
    │   │   ├── playlist.m3u8
    │   │   └── 1080p_001.ts
    │   └── 720p/
    │       ├── playlist.m3u8
    │       └── 720p_001.ts
    │
    ├── dash/                    <-- For the future
    │   └── ...
    │
    └── thumbnails/              <-- For the future
        └── thumbnail_01.jpg
```
