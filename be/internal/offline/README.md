# Offline Material Module

The Offline Material module provides DRM-like protection for downloadable educational materials in NgasihTau. It ensures content can only be accessed within the NgasihTau platform through device registration, license management, and content encryption.

## Architecture

```
internal/offline/
├── application/           # Business logic services
│   ├── device_service.go      # Device registration & management
│   ├── license_service.go     # License issuance & validation
│   ├── encryption_service.go  # Content encryption (AES-256-GCM)
│   ├── key_management_service.go  # CEK generation & management
│   ├── download_service.go    # Download preparation
│   ├── job_service.go         # Background job management
│   ├── rate_limiter.go        # Rate limiting logic
│   └── security_service.go    # Security features
├── domain/                # Core entities & interfaces
│   ├── entity.go              # Device, License, CEK, etc.
│   ├── repository.go          # Repository interfaces
│   ├── constants.go           # Domain constants
│   └── errors.go              # Domain-specific errors
├── infrastructure/        # External implementations
│   ├── postgres/              # PostgreSQL repositories
│   ├── redis/                 # Redis caching
│   ├── minio/                 # MinIO storage
│   └── metrics/               # Prometheus metrics
├── interfaces/
│   └── http/                  # HTTP handlers
│       ├── device_handler.go
│       ├── license_handler.go
│       ├── download_handler.go
│       ├── job_handler.go
│       └── health_handler.go
└── module.go              # Module initialization
```

## Features

### Device Management
- Register up to 5 devices per user
- Device fingerprint validation
- Automatic license revocation on device deregistration

### License Management
- Issue licenses for offline material access
- 30-day default expiration with 72-hour offline grace period
- License validation with nonce-based anti-cloning protection
- License renewal before expiration

### Content Encryption
- AES-256-GCM encryption for all materials
- 1MB chunk-based encryption for large files
- Per-chunk IV derivation for security
- Manifest generation with chunk metadata

### Security
- Rate limiting (10 downloads/hour per user)
- Device blocking after 5 failed validations
- Request replay protection (5-minute window)
- Audit logging for all security-sensitive operations

## API Endpoints

### Device Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/offline/devices` | Register a new device |
| GET | `/api/v1/offline/devices` | List user's devices |
| DELETE | `/api/v1/offline/devices/:device_id` | Deregister a device |

### License Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/offline/materials/:material_id/license` | Issue a license |
| POST | `/api/v1/offline/licenses/:license_id/validate` | Validate a license |
| POST | `/api/v1/offline/licenses/:license_id/renew` | Renew a license |

### Download Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/offline/materials/:material_id/download` | Download encrypted material |

### Job Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/offline/jobs/:job_id` | Get job status |
| GET | `/api/v1/offline/materials/:material_id/job` | Get job for material |
| POST | `/api/v1/offline/sync` | Sync offline state |

### Health Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe |
| GET | `/health/full` | Full health check |

## Configuration

Environment variables:

```bash
# Database
OFFLINE_DB_HOST=localhost
OFFLINE_DB_PORT=5432
OFFLINE_DB_NAME=ngasihtau_offline
OFFLINE_DB_USER=postgres
OFFLINE_DB_PASSWORD=secret

# Redis
OFFLINE_REDIS_HOST=localhost
OFFLINE_REDIS_PORT=6379

# MinIO
OFFLINE_MINIO_ENDPOINT=localhost:9000
OFFLINE_MINIO_ACCESS_KEY=minioadmin
OFFLINE_MINIO_SECRET_KEY=minioadmin
OFFLINE_MINIO_BUCKET=offline-materials

# Security
OFFLINE_MASTER_SECRET=your-32-byte-secret-key
OFFLINE_KEK_VERSION=1

# Rate Limiting
OFFLINE_DOWNLOAD_RATE_LIMIT=10
OFFLINE_DOWNLOAD_RATE_WINDOW=1h
OFFLINE_DEVICE_BLOCK_THRESHOLD=5
OFFLINE_DEVICE_BLOCK_DURATION=1h
```

## Database Migrations

Run migrations from the `be/` directory:

```bash
# Apply migrations
task migrate:offline:up

# Rollback migrations
task migrate:offline:down
```

Migration files are in `migrations/offline/`:
- `000001_create_offline_devices_table.up.sql`
- `000002_create_offline_ceks_table.up.sql`
- `000003_create_offline_licenses_table.up.sql`
- `000004_create_offline_encrypted_materials_table.up.sql`
- `000005_create_offline_audit_logs_table.up.sql`
- `000006_create_offline_device_rate_limits_table.up.sql`
- `000007_create_offline_encryption_jobs_table.up.sql`
- `000008_create_offline_sync_state_table.up.sql`

## Testing

```bash
# Run all offline module tests
task test -- ./internal/offline/...

# Run with coverage
task test:coverage -- ./internal/offline/...

# Run integration tests
go test -v -tags=integration ./internal/offline/...

# Run property-based tests
go test -v ./internal/offline/... -run "Property"
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `DEVICE_LIMIT_EXCEEDED` | 403 | User has 5 devices registered |
| `DEVICE_NOT_FOUND` | 404 | Device not registered or revoked |
| `DEVICE_FINGERPRINT_MISMATCH` | 403 | Fingerprint doesn't match |
| `DEVICE_BLOCKED` | 429 | Too many failed attempts |
| `LICENSE_NOT_FOUND` | 404 | License doesn't exist |
| `LICENSE_EXPIRED` | 403 | License past expiration |
| `LICENSE_REVOKED` | 403 | License was revoked |
| `LICENSE_OFFLINE_EXPIRED` | 403 | Offline grace period exceeded |
| `MATERIAL_ACCESS_DENIED` | 403 | No access to material |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many downloads |
| `INVALID_NONCE` | 403 | Nonce validation failed |

## Metrics

Prometheus metrics exposed at `/metrics`:

### Counters
- `offline_devices_registered_total` - Total devices registered
- `offline_licenses_issued_total` - Total licenses issued
- `offline_downloads_total` - Total downloads
- `offline_rate_limit_exceeded_total` - Rate limit violations

### Histograms
- `offline_encryption_duration_seconds` - Encryption operation duration
- `offline_download_duration_seconds` - Download preparation duration

### Gauges
- `offline_active_devices` - Currently active devices
- `offline_active_licenses` - Currently active licenses
- `offline_pending_jobs` - Pending encryption jobs

## Flow Diagrams

### Device Registration Flow
```
Client -> POST /offline/devices
       -> Validate fingerprint format
       -> Check device limit (max 5)
       -> Store device in DB
       -> Return device info
```

### License Issuance Flow
```
Client -> POST /offline/materials/:id/license
       -> Validate device ownership
       -> Check material access (via material-service)
       -> Generate/retrieve CEK
       -> Create license with nonce
       -> Return license + encrypted CEK
```

### Download Flow
```
Client -> GET /offline/materials/:id/download
       -> Validate license
       -> Check rate limits
       -> Get/create encrypted material
       -> Generate presigned URL
       -> Return manifest + download URL
```

## Dependencies

- **user-service**: Authentication and user validation
- **material-service**: Material access verification
- **PostgreSQL**: Primary data storage
- **Redis**: Caching and rate limiting
- **MinIO**: Encrypted file storage
- **NATS**: Event publishing
