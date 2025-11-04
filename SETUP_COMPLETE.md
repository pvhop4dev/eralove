# ✅ Setup Complete - Directus Integration

## Những gì đã hoàn thành

### 1. Infrastructure
- ✅ Thêm PostgreSQL (port 5432)
- ✅ Thêm Directus CMS (port 8055)
- ✅ Loại bỏ MongoDB
- ✅ Cấu hình Directus với memory cache
- ✅ Cập nhật docker-compose.yml và docker-compose.dev.yml
- ✅ Cập nhật nginx config để route Directus

### 2. Backend Code
- ✅ Tạo PostgreSQL client (`internal/infrastructure/database/postgres.go`)
- ✅ Tạo Directus client (`internal/infrastructure/directus/`)
- ✅ Tạo CMS service (`internal/service/cms_service.go`)
- ✅ Tạo CMS handler (`internal/handler/cms_handler.go`)
- ✅ Cập nhật Wire DI providers
- ✅ Regenerate Wire code
- ✅ Regenerate Swagger documentation
- ✅ Install PostgreSQL driver (`github.com/lib/pq`)

### 3. Configuration
- ✅ Cập nhật `.env.example` với PostgreSQL & Directus config
- ✅ Copy `.env.example` sang `.env`
- ✅ Cập nhật `config.go` với PostgreSQL & Directus fields

### 4. Documentation
- ✅ README.md
- ✅ QUICKSTART.md
- ✅ START_SERVICES.md
- ✅ MIGRATION_SUMMARY.md

### 5. Build Tools
- ✅ Makefile với cross-platform support (Windows + Ubuntu)
- ✅ Commands mới cho PostgreSQL & Directus

## 🚀 Khởi động Backend

```bash
# 1. Đảm bảo infrastructure đang chạy
make infra-up

# 2. Kiểm tra Directus đã sẵn sàng
docker logs eralove-directus-dev

# 3. Chạy backend
cd backend
go run cmd/main.go
```

Backend sẽ chạy trên: **http://localhost:8080**

## 🔑 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| **Backend API** | http://localhost:8080 | - |
| **Directus Admin** | http://localhost:8055 | admin@eralove.com / Admin@123456 |
| **Swagger Docs** | http://localhost:8080/swagger/ | - |
| **PostgreSQL** | localhost:5432 | directus / directus123 |

## 📡 CMS API Endpoints

### Public Endpoints
```bash
GET  /api/v1/cms/:collection          # Lấy items từ collection
GET  /api/v1/cms/:collection/:id      # Lấy item theo ID
GET  /api/v1/cms/blog/posts           # Lấy blog posts
GET  /api/v1/cms/pages                # Lấy pages
GET  /api/v1/cms/settings             # Lấy settings
GET  /api/v1/cms/files                # Lấy files
```

### Protected Endpoints (cần authentication)
```bash
POST   /api/v1/cms/:collection        # Tạo item mới
PATCH  /api/v1/cms/:collection/:id    # Cập nhật item
DELETE /api/v1/cms/:collection/:id    # Xóa item
```

## 🎯 Next Steps

### 1. Truy cập Directus Admin
```
URL: http://localhost:8055
Email: admin@eralove.com
Password: Admin@123456
```

### 2. Tạo Collection đầu tiên

Trong Directus Admin:
1. Vào **Settings → Data Model**
2. Click **Create Collection**
3. Tên: `posts`
4. Thêm fields:
   - `title` (String, Required)
   - `slug` (String, Required, Unique)
   - `content` (WYSIWYG)
   - `status` (Dropdown: draft, published)
   - `published_at` (DateTime)
5. Vào **Settings → Roles & Permissions**
6. Chọn **Public** role
7. Set permissions cho collection `posts`: Read only

### 3. Test API

```bash
# Test Directus health
curl http://localhost:8055/server/health

# Test backend health
curl http://localhost:8080/health

# Test CMS endpoint (sau khi tạo collection)
curl http://localhost:8080/api/v1/cms/posts
```

### 4. Sử dụng từ Frontend

```typescript
// Fetch posts từ Directus qua Go backend
const fetchPosts = async () => {
  const response = await fetch('http://localhost:8080/api/v1/cms/posts?limit=10');
  const data = await response.json();
  return data;
};

// Fetch specific post
const fetchPost = async (id: string) => {
  const response = await fetch(`http://localhost:8080/api/v1/cms/posts/${id}`);
  const post = await response.json();
  return post;
};
```

## 🛠️ Useful Commands

```bash
# Infrastructure
make infra-up              # Start all services
make infra-down            # Stop all services
make health                # Check service health
make status                # Show service status

# Database
make db-shell-postgres     # Open PostgreSQL shell
make directus-admin        # Show Directus credentials

# Logs
make logs-directus         # View Directus logs
make logs-postgres         # View PostgreSQL logs

# Backend
cd backend
go run cmd/main.go         # Run backend
go test ./...              # Run tests
wire gen ./internal/app    # Regenerate Wire DI
swag init -g cmd/main.go -o docs  # Regenerate Swagger
```

## ⚠️ Lưu ý

### MongoDB Legacy
- MongoDB providers vẫn tồn tại cho backward compatibility
- User và Photo repositories vẫn dùng MongoDB (nếu configured)
- Để migrate hoàn toàn sang PostgreSQL, cần update repositories

### Directus Cache
- Hiện tại dùng memory cache (không dùng Redis)
- Để enable Redis cache, update `docker-compose.dev.yml`:
  ```yaml
  CACHE_ENABLED: "true"
  CACHE_STORE: "redis"
  REDIS: "redis://:password123@redis:6379"
  ```

## 🐛 Troubleshooting

### Backend không start
```bash
# Check logs
cd backend
go run cmd/main.go

# Nếu lỗi MongoDB connection:
# - Kiểm tra .env file
# - Đảm bảo MONGO_URI="" (empty) để skip MongoDB
```

### Directus không accessible
```bash
# Check logs
docker logs eralove-directus-dev

# Restart
docker-compose -f docker-compose.dev.yml restart directus

# Wait 10-15 seconds then try: http://localhost:8055
```

### PostgreSQL connection failed
```bash
# Check logs
docker logs eralove-postgres-dev

# Test connection
make db-shell-postgres
```

## 📚 Documentation

- [README.md](./README.md) - Overview
- [QUICKSTART.md](./QUICKSTART.md) - Quick start guide
- [MIGRATION_SUMMARY.md](./MIGRATION_SUMMARY.md) - Migration details
- [Directus Docs](https://docs.directus.io)
- [Go Fiber Docs](https://docs.gofiber.io)

## ✨ Features

### Directus CMS
- ✅ Admin dashboard sẵn có
- ✅ Visual schema builder
- ✅ User & permission management
- ✅ File management với MinIO
- ✅ RESTful API tự động
- ✅ Filtering, sorting, pagination built-in

### Go Backend
- ✅ Wrap Directus API
- ✅ Business logic layer
- ✅ AI integration ready
- ✅ Authentication & Authorization
- ✅ File upload handling
- ✅ Swagger documentation

### Architecture
```
Frontend (React)
    ↓
Go Backend (API + AI Logic)
    ↓
Directus CMS (Admin Dashboard)
    ↓
PostgreSQL (Database)
```

## 🎉 Success!

Hệ thống đã sẵn sàng! Bạn có thể:
1. Truy cập Directus Admin để tạo collections
2. Chạy backend để test API
3. Tích hợp frontend với backend API
4. Thêm business logic và AI features vào Go backend
