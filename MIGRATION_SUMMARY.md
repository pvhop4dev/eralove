# Migration Summary: MongoDB → PostgreSQL + Directus

## Thay đổi chính

### 1. Database Migration
- ❌ **Removed**: MongoDB
- ✅ **Added**: PostgreSQL (port 5432)
- ✅ **Added**: Directus CMS (port 8055)

### 2. Kiến trúc mới

```
TRƯỚC:
Frontend → Go Backend → MongoDB
                     ↓
                  Redis + MinIO

SAU:
Frontend → Go Backend → Directus CMS → PostgreSQL
                     ↓
                  Redis + MinIO
```

### 3. Files đã tạo mới

#### Backend Infrastructure
```
backend/internal/infrastructure/
├── directus/
│   ├── client.go          # Directus API client
│   └── config.go          # Directus configuration
└── database/
    └── postgres.go        # PostgreSQL client
```

#### Backend Service & Handler
```
backend/internal/
├── service/
│   └── cms_service.go     # CMS business logic
├── handler/
│   └── cms_handler.go     # CMS HTTP handlers
└── app/
    ├── wire_providers.go  # DI providers
    └── routes/
        └── cms_routes.go  # CMS routes
```

#### Configuration
```
backend/
├── .env.example           # Updated với Directus config
└── internal/config/
    └── config.go          # Added PostgreSQL & Directus config
```

#### Docker & Infrastructure
```
├── docker-compose.yml     # Added PostgreSQL + Directus
├── docker-compose.dev.yml # Added PostgreSQL + Directus
└── nginx/conf.d/
    └── default.conf       # Added Directus routing
```

#### Documentation
```
├── README.md              # Updated với kiến trúc mới
├── QUICKSTART.md          # Quick start guide
├── START_SERVICES.md      # Service startup guide
└── MIGRATION_SUMMARY.md   # This file
```

#### Build Tools
```
└── Makefile               # Updated commands cho PostgreSQL/Directus
```

### 4. Environment Variables mới

```bash
# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=directus
POSTGRES_PASSWORD=directus123
POSTGRES_DB=directus
POSTGRES_SSLMODE=disable

# Directus
DIRECTUS_URL=http://localhost:8055
DIRECTUS_ADMIN_EMAIL=admin@eralove.com
DIRECTUS_ADMIN_PASSWORD=Admin@123456
DIRECTUS_STATIC_TOKEN=
```

### 5. API Endpoints mới

```
# CMS Public Endpoints
GET  /api/v1/cms/:collection
GET  /api/v1/cms/:collection/:id
GET  /api/v1/cms/blog/posts
GET  /api/v1/cms/pages
GET  /api/v1/cms/settings
GET  /api/v1/cms/files

# CMS Protected Endpoints (cần auth)
POST   /api/v1/cms/:collection
PATCH  /api/v1/cms/:collection/:id
DELETE /api/v1/cms/:collection/:id
```

### 6. Docker Services

#### Removed
- ❌ `mongodb` container
- ❌ `mongodb_data` volume

#### Added
- ✅ `postgres` container (port 5432)
- ✅ `directus` container (port 8055)
- ✅ `postgres_data` volume
- ✅ `directus_uploads` volume
- ✅ `directus_extensions` volume

### 7. Makefile Commands mới

```bash
# Database
make db-shell-postgres     # PostgreSQL shell (thay vì mongo)
make directus-admin        # Show Directus credentials

# Logs
make logs-postgres         # PostgreSQL logs
make logs-directus         # Directus logs

# Health checks updated
make health                # Bao gồm PostgreSQL + Directus
```

## Lợi ích của Migration

### 1. Không cần code CRUD
- Directus cung cấp admin dashboard sẵn
- Tạo/sửa/xóa data qua UI
- Không cần viết repository/handler cho mỗi model

### 2. Schema Management
- Quản lý database schema qua Directus UI
- Không cần viết migration scripts
- Visual schema builder

### 3. User & Permission Management
- Directus có sẵn user system
- Role-based access control
- Không cần implement từ đầu

### 4. File Management
- Upload files qua Directus
- Tích hợp với MinIO
- Thumbnail generation tự động

### 5. RESTful API tự động
- Directus tự động generate API cho mọi collection
- Filtering, sorting, pagination built-in
- GraphQL support (nếu cần)

### 6. Go Backend Focus
- Focus vào business logic
- AI integration
- Custom endpoints
- Không mất thời gian cho CRUD

## Next Steps

### 1. Install Dependencies

```bash
cd backend
go mod tidy
```

### 2. Start Services

```bash
make infra-up
```

### 3. Access Directus Admin

- URL: http://localhost:8055
- Email: admin@eralove.com
- Password: Admin@123456

### 4. Tạo Collections

Trong Directus Admin:
1. Settings → Data Model
2. Create Collection
3. Add Fields
4. Set Permissions

### 5. Test API

```bash
# Test Directus health
curl http://localhost:8055/server/health

# Test CMS endpoint (qua Go backend)
curl http://localhost:8080/api/v1/cms/settings
```

## Migration Checklist

- [x] Thêm PostgreSQL vào docker-compose
- [x] Thêm Directus vào docker-compose
- [x] Tạo Directus client trong Go
- [x] Tạo CMS service layer
- [x] Tạo CMS handlers
- [x] Cập nhật routes
- [x] Cập nhật nginx config
- [x] Cập nhật .env.example
- [x] Cập nhật Makefile
- [x] Cập nhật documentation
- [ ] Run `go mod tidy` để install PostgreSQL driver
- [ ] Test infrastructure startup
- [ ] Test API endpoints
- [ ] Migrate existing data (if any)

## Troubleshooting

### Import error: github.com/lib/pq

```bash
cd backend
go mod tidy
```

### Directus không khởi động

```bash
docker logs eralove-directus-dev
docker-compose -f docker-compose.dev.yml restart directus
```

### PostgreSQL connection failed

```bash
docker logs eralove-postgres-dev
make db-shell-postgres
```

## References

- Directus Documentation: https://docs.directus.io
- PostgreSQL Documentation: https://www.postgresql.org/docs/
- Go pq driver: https://github.com/lib/pq
