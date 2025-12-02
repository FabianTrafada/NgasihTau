# File Processor Service

A lightweight Python service for extracting text from PDF, DOCX, and PPTX files. Part of the NgasihTau backend microservices architecture.

## Features

- Extract text from PDF files (using PyMuPDF)
- Extract text from DOCX files (using python-docx)
- Extract text from PPTX files (using python-pptx)
- Returns metadata including page count and word count
- Health check endpoint for container orchestration

## Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/) - Fast Python package manager

## Development Setup

1. **Install uv** (if not already installed):
   ```bash
   # Using pip
   pip install uv

   # Or using Homebrew (macOS)
   brew install uv

   # Or using the installer script
   curl -LsSf https://astral.sh/uv/install.sh | sh
   ```

2. **Create virtual environment and install dependencies**:
   ```bash
   cd be/file-processor
   uv venv
   uv sync
   ```

3. **Activate the virtual environment**:
   ```bash
   source .venv/bin/activate
   ```

4. **Run the service**:
   ```bash
   uvicorn main:app --reload --port 8000
   ```

## API Endpoints

### Health Check
```
GET /health
```

Returns the service health status.

**Response:**
```json
{
  "status": "healthy",
  "service": "file-processor",
  "timestamp": "2025-12-02T10:00:00Z"
}
```

### Extract Text
```
POST /extract
```

Extracts text content from a file.

**Request Body:**
```json
{
  "file_url": "http://minio:9000/bucket/file.pdf",
  "file_type": "pdf"
}
```

**Response:**
```json
{
  "text": "Extracted text content...",
  "metadata": {
    "pages": 10,
    "word_count": 5000,
    "char_count": 25000
  }
}
```

## Docker

### Build the image
```bash
docker build -t file-processor .
```

### Run the container
```bash
docker run -p 8000:8000 file-processor
```

## Testing

```bash
# Install dev dependencies
uv sync --all-extras

# Run tests
pytest

# Run tests with verbose output
pytest -v

# Run specific test file
pytest tests/test_main.py
```

## Development Workflow

### Code Quality

```bash
# Format code with ruff
ruff format .

# Lint code
ruff check .

# Fix auto-fixable lint issues
ruff check --fix .
```

### Dependency Management

```bash
# Add a new dependency
uv add <package-name>

# Add a dev dependency
uv add --optional dev <package-name>

# Update dependencies
uv lock --upgrade

# Sync dependencies from lock file (reproducible builds)
uv sync
```

### Reproducible Builds

The `uv.lock` file ensures reproducible builds across all environments. Always commit this file to version control.

```bash
# Install exact versions from lock file
uv sync

# Install with dev dependencies
uv sync --all-extras
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HOST` | Server host | `0.0.0.0` |
| `PORT` | Server port | `8000` |

## License

Part of the NgasihTau project.
