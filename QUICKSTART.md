# EraLove - Quick Start Guide

## Kiến trúc mới với Directus CMS

```
┌─────────────┐
│   Frontend  │  React (Port 5173)
│   (React)   │
└──────┬──────┘
       │
       ↓
┌──────────────┐
│  Go Backend  │  API Server (Port 8080)
│   + AI Logic │
└──────┬───────┘
       │
       ├──→ Directus CMS (Port 8055) ──→ PostgreSQL (Port 5432)
       ├──→ Redis Cache (Port 6379)
       └──→ MinIO Storage (Port 9000)
```

## Khởi động nhanh (3 bước)

### 1. Start Infrastructure

```bash
make infra-up
```

Lệnh này sẽ khởi động:
- ✅ PostgreSQL (database chính)
- ✅ Directus CMS (admin dashboard)
- ✅ Redis (cache)
- ✅ MinIO (object storage)
- ✅ Nginx (reverse proxy)

### 2. Start Backend

```bash
cd backend
go mod tidy
go run cmd/main.go
```

### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

## Truy cập Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Frontend** | http://localhost:5173 | - |
| **Backend API** | http://localhost:8080 | - |
| **Directus Admin** | http://localhost:8055 | admin@eralove.com / Admin@123456 |
| **MinIO Console** | http://localhost:9001 | minioadmin / minioadmin123 |
| **Swagger Docs** | http://localhost:8080/swagger/ | - |

## Sử dụng Directus CMS

### Tạo Collection đầu tiên

1. Mở Directus Admin: http://localhost:8055
2. Login với credentials trên
3. Vào **Settings → Data Model**
4. Click **Create Collection**
5. Tạo collection `posts` với các fields:
   - `title` (String, Required)
   - `content` (WYSIWYG)
   - `status` (Dropdown: draft, published)
   - `published_at` (DateTime)

### Gọi API từ Frontend

```typescript
// Lấy tất cả posts
const response = await fetch('http://localhost:8080/api/v1/cms/posts');
const posts = await response.json();

// Lấy post theo ID
const response = await fetch('http://localhost:8080/api/v1/cms/posts/1');
const post = await response.json();
```

## API Endpoints

### CMS Endpoints (Public)

```
GET  /api/v1/cms/:collection          # Lấy tất cả items
GET  /api/v1/cms/:collection/:id      # Lấy item theo ID
GET  /api/v1/cms/blog/posts           # Lấy blog posts
GET  /api/v1/cms/pages                # Lấy pages
GET  /api/v1/cms/settings             # Lấy settings
```

### CMS Endpoints (Protected - cần auth)

```
POST   /api/v1/cms/:collection        # Tạo item mới
PATCH  /api/v1/cms/:collection/:id    # Cập nhật item
DELETE /api/v1/cms/:collection/:id    # Xóa item
```

## Makefile Commands

```bash
make help              # Xem tất cả commands
make infra-up          # Start infrastructure
make infra-down        # Stop infrastructure
make backend           # Start Go backend
make frontend          # Start React frontend
make health            # Check service health
make status            # Show service status
make directus-admin    # Show Directus credentials
make db-shell-postgres # Open PostgreSQL shell
make logs-directus     # View Directus logs
```

## Troubleshooting

### PostgreSQL không khởi động
```bash
docker logs eralove-postgres-dev
make logs-postgres
```

### Directus không khởi động
```bash
docker logs eralove-directus-dev
make logs-directus

# Restart
docker-compose -f docker-compose.dev.yml restart directus
```

### Backend không kết nối được Directus
- Kiểm tra Directus đã chạy: `curl http://localhost:8055/server/health`
- Kiểm tra `.env` file trong backend folder
- Đảm bảo `DIRECTUS_URL=http://localhost:8055`

### Reset toàn bộ database
```bash
make db-reset  # ⚠️ Cẩn thận: Sẽ xóa tất cả dữ liệu
```

## Lợi ích của kiến trúc này

✅ **Không cần code CRUD** - Directus cung cấp admin dashboard sẵn
✅ **PostgreSQL duy nhất** - Không cần MongoDB
✅ **Frontend gọi qua Go** - Thêm logic nghiệp vụ và AI processing
✅ **Schema management** - Quản lý database schema qua UI
✅ **User management** - Directus có sẵn user & permission system
✅ **File management** - Upload và quản lý files qua Directus
✅ **RESTful API** - Directus tự động generate API endpoints

## Next Steps

1. Tạo collections trong Directus cho data models của bạn
2. Set permissions cho public/authenticated users
3. Implement business logic trong Go backend
4. Tích hợp AI features trong Go backend
5. Build UI trong React frontend

## Support

- Directus Docs: https://docs.directus.io
- PostgreSQL Docs: https://www.postgresql.org/docs/
- Fiber (Go) Docs: https://docs.gofiber.io
