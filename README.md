# NgasihTau

**NgasihTau** adalah platform berbagi pengetahuan yang terdiri dari beberapa microservice Go (backend) dan frontend Next.js. Repository ini berisi definisi infrastruktur, skrip migrasi, dan task helper untuk memudahkan pengembangan dan pengujian lokal.

---

## 🚀 Mulai Cepat

### Persyaratan

- Go 1.21+
- Node.js (untuk frontend)
- npm / yarn / pnpm
- Docker & Docker Compose
- Task (Taskfile) - digunakan untuk perintah pengembangan (install: `go install github.com/go-task/task/v3/cmd/task@latest`)
- golang-migrate (untuk migrasi DB) - opsional untuk migrasi lokal

### 1) Jalankan infrastruktur

Dari folder `be/` (atau root repo jika perlu), jalankan infrastruktur yang diperlukan (Postgres, Redis, MinIO, NATS, Meilisearch, Qdrant, Traefik, dll.):

```bash
cd be
task infra:up
```

Untuk menghentikan infrastruktur:

```bash
task infra:down
```

### 2) Siapkan environment

Salin file contoh `.env` dan ubah jika perlu:

```bash
cd be
cp .env.example .env
```

### 3) Jalankan migrasi database (contoh untuk service user)

```bash
task migrate:user:up
```

### 4) Jalankan layanan backend

Jalankan layanan di terminal terpisah. Contoh:

```bash
# User Service (port 8001)
task dev:user

# Pod Service (port 8002)
task dev:pod

# Material Service (port 8003)
task dev:material
```

Untuk melihat daftar perintah `task`:

```bash
task --list
```

### 5) Jalankan frontend (Next.js)

Dari `fe/`:

```bash
cd fe
npm install
npm run dev
# atau: pnpm dev / yarn dev
```

Buka http://localhost:3000

---

## 🧩 Struktur Proyek (sekilas)

- `be/` - Backend microservices ditulis dengan Go
  - `cmd/` - Entrypoint utama untuk tiap layanan (user, pod, material, ai, notification, search)
  - `internal/` - Internal layanan (domain, application, infra, interfaces)
  - `migrations/` - Migrasi SQL untuk tiap layanan
  - `api/swagger/` - Spesifikasi Swagger/OpenAPI
  - `Taskfile.yml` - Task dan skrip pengembangan
  - `.env.example` - Contoh variabel environment
- `fe/` - Frontend (Next.js, TypeScript)
- `file-processor/` - Utilitas pemrosesan file dengan Python
- `postman/` - Koleksi Postman (`NgasihTau_API.postman_collection.json`)
- `traefik/` - Konfigurasi Traefik dan aturan dinamis

---

## 🔧 Perintah Berguna

Backend (dari `be/`):

- `task infra:up` - Menjalankan infrastruktur
- `task infra:down` - Menghentikan infrastruktur
- `task dev:user` - Menjalankan User service
- `task dev:pod` - Menjalankan Pod service
- `task dev:material` - Menjalankan Material service
- `task migrate:user:up` - Menjalankan migrasi untuk service user
- `task test` - Menjalankan tes
- `task lint`, `task fmt`, `task vet` - Alat kualitas kode

Frontend (dari `fe/`):

- `npm install` - Menginstal dependensi
- `npm run dev` - Menjalankan dev server (http://localhost:3000)
- `npm run build` - Membangun untuk produksi
- `npm run start` - Menjalankan aplikasi hasil build

Python file-processor (di `file-processor/`):

- `python main.py` (lihat README di `file-processor/` untuk detail)

---

## 📡 Layanan & Port (default pengembangan)

- User service - 8001
- Pod service - 8002
- Material service - 8003
- Search service - 8004
- AI service - 8005
- Notification service - 8006
- PostgreSQL - 5432
- Redis - 6379
- MinIO API - 9000, Console - 9001
- Meilisearch - 7700
- Qdrant - 6333
- Traefik Dashboard - 8081

---

## 🧪 Pengujian & Eksplorasi API

- Postman: `postman/NgasihTau_API.postman_collection.json`
- Spesifikasi Swagger/OpenAPI tersedia di `be/api/swagger/` dan tiap layanan di `be/cmd/*/`

Contoh: Daftar pengguna

```http
POST http://localhost:8001/api/v1/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "SecurePass123!",
  "name": "Test User"
}
```

Login akan mengembalikan `access_token` yang digunakan di header `Authorization: Bearer <token>` untuk endpoint yang dilindungi.

---

## 🛠️ Kontribusi

- Jalankan `task lint` dan `task fmt` sebelum mengirim PR untuk backend.
- Tambahkan tes untuk perilaku baru atau yang diubah (`task test`).
- Ikuti batasan layanan: logika bisnis berada di `internal/<service>/domain` dan orkestrasi di `internal/<service>/application`.

---

## ⚠️ Pemecahan Masalah

- Jika layanan gagal mulai, pastikan infrastruktur berjalan: `task infra:ps` / `task infra:up`.
- Periksa nilai `.env` dan pastikan port yang diperlukan tersedia.
- Periksa log tiap layanan saat menggunakan perintah `task` atau log Docker Compose.

---

## 📚 Referensi

- Backend README: `be/README.md`
- Frontend README: `fe/README.md`
- Koleksi Postman: `postman/NgasihTau_API.postman_collection.json`
- Spesifikasi Swagger: `be/api/swagger/`
