# NgasihTau - VPS Deployment Guide (Docker)

Panduan step-by-step untuk deploy NgasihTau di VPS menggunakan Docker Compose.

## Prerequisites

- VPS dengan minimal 4GB RAM (recommended 8GB)
- Ubuntu 22.04 LTS
- Domain yang sudah pointing ke IP VPS
- Port 80 dan 443 terbuka

## Step 1: Setup VPS

SSH ke VPS kamu:

```bash
ssh root@YOUR_VPS_IP
```

Update system:

```bash
apt update && apt upgrade -y
```

## Step 2: Install Docker

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Add user ke docker group (jika bukan root)
usermod -aG docker $USER

# Install Docker Compose plugin
apt install docker-compose-plugin -y

# Verify
docker --version
docker compose version
```

## Step 3: Setup DNS

Di domain provider (Cloudflare/Namecheap/dll), buat A records:

| Type | Name | Value |
|------|------|-------|
| A | @ | YOUR_VPS_IP |
| A | api | YOUR_VPS_IP |
| A | www | YOUR_VPS_IP |

Tunggu propagasi DNS (5-30 menit).

## Step 4: Clone Repository

```bash
# Buat directory
mkdir -p /opt/ngasihtau
cd /opt/ngasihtau

# Clone repo
git clone https://github.com/YOUR_USERNAME/ngasihtau.git .

# Atau upload manual via scp/sftp
```

## Step 5: Setup Environment

```bash
# Copy template
cp .env.production.example .env

# Edit dengan nano/vim
nano .env
```

**WAJIB diganti:**
- `DOMAIN` - domain kamu
- `NEXT_PUBLIC_API_URL` - https://api.yourdomain.com
- `NEXT_PUBLIC_APP_URL` - https://yourdomain.com
- Semua `PASSWORD` - generate dengan `openssl rand -base64 32`
- `JWT_SECRET` - minimal 32 karakter
- `OPENAI_API_KEY` - API key OpenAI kamu
- `CORS_ALLOWED_ORIGINS` - domain kamu

## Step 6: Setup SSL Certificate

### Option A: Self-signed (untuk testing)

```bash
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/privkey.pem \
  -out nginx/ssl/fullchain.pem \
  -subj "/CN=yourdomain.com"
```

### Option B: Let's Encrypt (untuk production)

```bash
# Install certbot
apt install certbot -y

# Stop nginx dulu jika running
docker compose -f docker-compose.prod.yml down nginx 2>/dev/null || true

# Generate certificate
certbot certonly --standalone -d yourdomain.com -d api.yourdomain.com -d www.yourdomain.com

# Copy ke nginx folder
mkdir -p nginx/ssl
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
```

## Step 7: Update Nginx Config

Edit `nginx/nginx.conf` dan ganti semua `yourdomain.com` dengan domain kamu:

```bash
sed -i 's/yourdomain.com/YOUR_ACTUAL_DOMAIN/g' nginx/nginx.conf
```

## Step 8: Make Scripts Executable

```bash
chmod +x be/scripts/init-databases.sh
```

## Step 9: Build & Deploy

```bash
# Build semua images (pertama kali akan lama ~10-15 menit)
docker compose -f docker-compose.prod.yml build

# Start semua services
docker compose -f docker-compose.prod.yml up -d

# Cek status
docker compose -f docker-compose.prod.yml ps
```

## Step 10: Run Migrations

Tunggu sampai semua services healthy (~1-2 menit), lalu:

```bash
# Masuk ke container postgres
docker compose -f docker-compose.prod.yml exec postgres bash

# Di dalam container, jalankan migrations
# User service
psql -U postgres -d ngasihtau_users -f /path/to/migrations/user/*.up.sql

# Atau keluar dan jalankan dari host:
exit
```

Untuk menjalankan migrations dari host:

```bash
# Copy migrations ke container
docker cp be/migrations/user postgres:/tmp/user_migrations
docker cp be/migrations/pod postgres:/tmp/pod_migrations
docker cp be/migrations/material postgres:/tmp/material_migrations
docker cp be/migrations/ai postgres:/tmp/ai_migrations
docker cp be/migrations/notification postgres:/tmp/notification_migrations

# Run migrations
docker compose -f docker-compose.prod.yml exec postgres bash -c "
  for f in /tmp/user_migrations/*.up.sql; do psql -U postgres -d ngasihtau_users -f \$f; done
  for f in /tmp/pod_migrations/*.up.sql; do psql -U postgres -d ngasihtau_pods -f \$f; done
  for f in /tmp/material_migrations/*.up.sql; do psql -U postgres -d ngasihtau_materials -f \$f; done
  for f in /tmp/ai_migrations/*.up.sql; do psql -U postgres -d ngasihtau_ai -f \$f; done
  for f in /tmp/notification_migrations/*.up.sql; do psql -U postgres -d ngasihtau_notifications -f \$f; done
"
```

## Step 11: Verify Deployment

```bash
# Cek semua services running
docker compose -f docker-compose.prod.yml ps

# Cek logs jika ada error
docker compose -f docker-compose.prod.yml logs -f

# Test API
curl https://api.yourdomain.com/api/v1/users/health

# Test frontend
curl https://yourdomain.com
```

## Useful Commands

```bash
# Lihat logs service tertentu
docker compose -f docker-compose.prod.yml logs -f user-service

# Restart service
docker compose -f docker-compose.prod.yml restart user-service

# Stop semua
docker compose -f docker-compose.prod.yml down

# Stop dan hapus volumes (HATI-HATI: data hilang!)
docker compose -f docker-compose.prod.yml down -v

# Update deployment
git pull
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d
```

## Auto-Renew SSL (Let's Encrypt)

Buat cron job untuk auto-renew:

```bash
# Edit crontab
crontab -e

# Tambahkan line ini (renew setiap hari jam 3 pagi)
0 3 * * * certbot renew --quiet && cp /etc/letsencrypt/live/yourdomain.com/*.pem /opt/ngasihtau/nginx/ssl/ && docker compose -f /opt/ngasihtau/docker-compose.prod.yml restart nginx
```

## Troubleshooting

### Service tidak start
```bash
docker compose -f docker-compose.prod.yml logs SERVICE_NAME
```

### Database connection error
Pastikan postgres sudah healthy:
```bash
docker compose -f docker-compose.prod.yml ps postgres
```

### SSL error
Cek certificate:
```bash
ls -la nginx/ssl/
openssl x509 -in nginx/ssl/fullchain.pem -text -noout
```

### Port already in use
```bash
# Cek port yang dipakai
netstat -tlnp | grep -E ':(80|443)'

# Stop service yang pakai port tersebut
systemctl stop apache2  # atau nginx jika ada
```

## Architecture

```
Internet
    │
    ▼
┌─────────────────────────────────────────┐
│              Nginx (80/443)             │
│         Reverse Proxy + SSL             │
└────────────────┬────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
Frontend    API Services   MinIO Console
 :3000      :8001-8006       :9001
    │            │
    └────────────┼────────────┐
                 │            │
    ┌────────────┼────────────┼────────────┐
    │            │            │            │
    ▼            ▼            ▼            ▼
PostgreSQL    Redis        NATS        MinIO
  :5432       :6379       :4222       :9000
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
Meilisearch   Qdrant    File Processor
  :7700       :6333        :8086
```
