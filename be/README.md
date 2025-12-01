# NgasihTau Backend

Backend services for NgasihTau - a knowledge sharing platform built with Go microservices.

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- [Task](https://taskfile.dev/) - Task runner (install: `go install github.com/go-task/task/v3/cmd/task@latest`)
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations (install: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

## Quick Start

### 1. Start Infrastructure

Start all required services (PostgreSQL, Redis, NATS, MinIO, etc.):

```bash
task infra:up
```

### 2. Setup Environment

Copy the example environment file:

```bash
cp .env.example .env
```

The default values work for local development. No changes needed.

### 3. Run Database Migrations

```bash
task migrate:user:up
```

### 4. Start Services

Run individual services in separate terminals:

```bash
# User Service (Authentication, Profiles) - Port 8001
task dev:user

# Pod Service (Knowledge Pods, Collaboration) - Port 8002
task dev:pod

# Material Service (File uploads) - Port 8003
task dev:material
```

## Available Commands

Run `task` or `task --list` to see all available commands.

### Development

| Command | Description |
|---------|-------------|
| `task dev:user` | Run User Service (port 8001) |
| `task dev:pod` | Run Pod Service (port 8002) |
| `task dev:material` | Run Material Service (port 8003) |
| `task dev:search` | Run Search Service (port 8004) |
| `task dev:ai` | Run AI Service (port 8005) |
| `task dev:notification` | Run Notification Service (port 8006) |

### Infrastructure

| Command | Description |
|---------|-------------|
| `task infra:up` | Start all infrastructure (Docker) |
| `task infra:down` | Stop all infrastructure |
| `task infra:logs` | View infrastructure logs |
| `task infra:ps` | Show infrastructure status |

### Database Migrations

| Command | Description |
|---------|-------------|
| `task migrate:user:up` | Run User Service migrations |
| `task migrate:user:down` | Rollback last migration |

### Testing

| Command | Description |
|---------|-------------|
| `task test` | Run all tests |
| `task test:unit` | Run unit tests only |
| `task test:coverage` | Run tests with coverage report |

### Code Quality

| Command | Description |
|---------|-------------|
| `task lint` | Run linter |
| `task fmt` | Format code |
| `task vet` | Run go vet |

## API Endpoints

### User Service (Port 8001)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/users/me` | Get current user profile |
| PUT | `/api/v1/users/me` | Update current user profile |

### Pod Service (Port 8002)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/pods` | Create new pod |
| GET | `/api/v1/pods` | List pods |
| GET | `/api/v1/pods/:id` | Get pod by ID |
| PUT | `/api/v1/pods/:id` | Update pod |
| DELETE | `/api/v1/pods/:id` | Delete pod |
| POST | `/api/v1/pods/:id/fork` | Fork a pod |
| POST | `/api/v1/pods/:id/star` | Star a pod |
| DELETE | `/api/v1/pods/:id/star` | Unstar a pod |

## Testing with Postman

### Register User

```
POST http://localhost:8001/api/v1/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "SecurePass123!",
  "name": "Test User"
}
```

### Login

```
POST http://localhost:8001/api/v1/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```

Response will include `access_token` - use it in Authorization header for protected endpoints:

```
Authorization: Bearer <access_token>
```

### Create Pod (requires auth)

```
POST http://localhost:8002/api/v1/pods
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My Learning Pod",
  "description": "A pod for learning materials",
  "visibility": "public",
  "categories": ["math", "science"],
  "tags": ["beginner"]
}
```

## Infrastructure Services

| Service | Port | Description |
|---------|------|-------------|
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |
| NATS | 4222 | Message queue |
| MinIO | 9000 (API), 9001 (Console) | Object storage |
| Meilisearch | 7700 | Full-text search |
| Qdrant | 6333 | Vector database |
| Traefik | 8000 (HTTP), 8081 (Dashboard) | API Gateway |

## Troubleshooting

### Services won't start

1. Make sure infrastructure is running: `task infra:ps`
2. Check if `.env` file exists
3. Run migrations: `task migrate:user:up`

### Database connection errors

1. Check PostgreSQL is healthy: `docker logs ngasihtau-postgres`
2. Verify database credentials in `.env`

### Port already in use

Check what's using the port and kill it:
```bash
lsof -i :8001
kill -9 <PID>
```
