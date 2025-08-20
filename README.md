# Better Media

Better Media is an opinionated way to build media software infrastructure using FFmpeg and S3. It is designed to work with all S3-compatible storage, and powered by asynq for transcoding queues such that the process can be distributed across multiple machines.

> [!WARNING]
> This project is a work in progress, and is not yet ready for production use.

## Running the Project

```bash
air -c .air.worker.toml
```

```bash
air -c .air.ingest.toml
```

## Roadmap

- [ ] Video on demand (ingest, encoding, storage, playback)
- [ ] Live streaming (ingest, encoding, storage, playback)
- [ ] Media management (metadata, thumbnails, etc.)
- [ ] Captions
- [ ] Analytics (views, engagement, etc.)
- [ ] User management (authentication, authorization, etc.)
- [ ] Admin dashboard (monitoring, management, etc.)
- [ ] Digital Rights Management (DRM)
