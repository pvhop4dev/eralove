# Hướng dẫn Khởi động EraLove với Directus

## Kiến trúc mới

```
Frontend (React) → Go Backend → Directus CMS → PostgreSQL
                              ↓
                          Redis Cache
                              ↓
                          MinIO Storage
```

## Khởi động Services

### 1. Start Infrastructure (Docker)

```bash
# Development mode
docker-compose -f docker-compose.dev.yml up -d

# Services được khởi động:
# - PostgreSQL (port 5432) - Database chính
# - Directus (port 8055) - CMS Admin & API
# - Redis (port 6379) - Cache
# - MinIO (port 9000, 9001) - Object Storage
# - Nginx (port 80) - Reverse Proxy
```

### 2. Cài đặt Go Dependencies

```bash
cd backend
go mod download
go mod tidy
```

### 3. Start Go Backend

```bash
cd backend
go run cmd/main.go

# Backend sẽ chạy trên port 8080
```

### 4. Start React Frontend

```bash
cd frontend
npm install
npm run dev

# Frontend sẽ chạy trên port 5173 hoặc 3000
```

## Truy cập Services

- **Frontend**: http://localhost:5173
- **Go Backend API**: http://localhost:8080
- **Directus Admin**: http://localhost:8055
  - Email: admin@eralove.com
  - Password: Admin@123456
- **MinIO Console**: http://localhost:9001
  - Username: minioadmin
  - Password: minioadmin123

## API Endpoints

### CMS Endpoints (qua Go Backend)

```
# Public endpoints
GET  /api/v1/cms/blog/posts          # Lấy blog posts
GET  /api/v1/cms/pages                # Lấy pages
GET  /api/v1/cms/settings             # Lấy settings
GET  /api/v1/cms/files                # Lấy files
GET  /api/v1/cms/:collection          # Lấy items từ collection
GET  /api/v1/cms/:collection/:id      # Lấy item theo ID

# Protected endpoints (cần authentication)
POST   /api/v1/cms/:collection        # Tạo item mới
PATCH  /api/v1/cms/:collection/:id    # Cập nhật item
DELETE /api/v1/cms/:collection/:id    # Xóa item
```

## Tạo Collections trong Directus

### Ví dụ: Blog Posts

1. Truy cập Directus Admin: http://localhost:8055
2. Login với admin credentials
3. Vào **Settings → Data Model**
4. Click **Create Collection**
5. Tên collection: `posts`
6. Thêm fields:
   - `title` (String, Required)
   - `slug` (String, Required, Unique)
   - `content` (WYSIWYG)
   - `excerpt` (Text)
   - `featured_image` (Image)
   - `status` (Dropdown: draft, published)
   - `published_at` (DateTime)
7. Set permissions cho **Public** role (read only)

### Ví dụ: Pages

```
Collection: pages
Fields:
- title (String)
- slug (String, Unique)
- content (WYSIWYG)
- meta_title (String)
- meta_description (Text)
```

### Ví dụ: Settings (Singleton)

```
Collection: settings
Type: Singleton
Fields:
- site_name (String)
- site_description (Text)
- logo (Image)
- contact_email (String)
```

## Sử dụng CMS từ Frontend

```typescript
// Example: Fetch blog posts
const response = await fetch('http://localhost:8080/api/v1/cms/blog/posts?limit=10');
const data = await response.json();

// Example: Fetch specific page
const response = await fetch('http://localhost:8080/api/v1/cms/pages/about');
const page = await response.json();

// Example: Fetch settings
const response = await fetch('http://localhost:8080/api/v1/cms/settings');
const settings = await response.json();
```

## Dừng Services

```bash
# Stop all Docker services
docker-compose -f docker-compose.dev.yml down

# Stop và xóa volumes (cẩn thận - sẽ mất dữ liệu)
docker-compose -f docker-compose.dev.yml down -v
```

## Troubleshooting

### PostgreSQL connection failed
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs eralove-postgres-dev
```

### Directus không khởi động
```bash
# Check logs
docker logs eralove-directus-dev

# Restart service
docker-compose -f docker-compose.dev.yml restart directus
```

### Go backend không kết nối được Directus
- Kiểm tra `DIRECTUS_URL` trong `.env`
- Đảm bảo Directus đã khởi động hoàn toàn
- Check logs: `docker logs eralove-directus-dev`
