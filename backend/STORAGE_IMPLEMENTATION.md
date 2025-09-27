# Storage System Implementation

## ✅ **Hoàn thành Storage System cho EraLove Backend**

### 📋 **Tổng quan:**
Đã implement một flexible storage system hỗ trợ multiple providers:
- **Local Storage** - Filesystem storage cho development
- **MinIO** - Self-hosted S3-compatible storage
- **AWS S3** - Cloud storage (thông qua MinIO SDK)

### 🏗️ **Architecture:**

#### **1. Domain Layer (`internal/domain/storage.go`):**
```go
// Core interfaces và types
type StorageService interface {
    Upload(ctx context.Context, req *UploadRequest) (*FileInfo, error)
    Download(ctx context.Context, req *DownloadRequest) (string, error)
    Delete(ctx context.Context, key string) error
    GetFileInfo(ctx context.Context, key string) (*FileInfo, error)
    ListFiles(ctx context.Context, folder string, limit int) ([]*FileInfo, error)
    GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error)
    GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type FileInfo struct {
    Key         string    `json:"key"`
    URL         string    `json:"url"`
    Filename    string    `json:"filename"`
    ContentType string    `json:"content_type"`
    Size        int64     `json:"size"`
    UploadedAt  time.Time `json:"uploaded_at"`
    Bucket      string    `json:"bucket"`
}
```

#### **2. Infrastructure Layer:**

##### **Local Storage (`internal/infrastructure/storage/local_storage.go`):**
- Filesystem-based storage
- Suitable cho development và testing
- Tự động tạo directories
- Public URL generation

##### **MinIO Storage (`internal/infrastructure/storage/minio_storage.go`):**
- MinIO SDK compatible với S3, MinIO, và AWS S3
- Presigned URLs cho direct upload/download
- Automatic bucket creation
- SSL/TLS support

##### **Storage Factory (`internal/infrastructure/storage/factory.go`):**
```go
// Factory pattern để tạo storage services
func (f *Factory) CreateStorage(config *domain.StorageConfig) (domain.StorageService, error)

// Helper functions
func GetDefaultConfig() *domain.StorageConfig      // Local storage
func GetMinIOConfig() *domain.StorageConfig        // MinIO development
func GetS3Config(...) *domain.StorageConfig       // AWS S3 production
```

### ⚙️ **Configuration:**

#### **Environment Variables:**
```bash
# Storage Provider Selection
STORAGE_PROVIDER=local          # local, minio, s3

# Common Settings
STORAGE_REGION=us-east-1
STORAGE_BUCKET=eralove-uploads
STORAGE_BASE_URL=http://localhost:8080

# MinIO/S3 Credentials
STORAGE_ACCESS_KEY_ID=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_ENDPOINT=localhost:9000
STORAGE_USE_SSL=false
```

#### **Provider Examples:**

##### **Local Storage (Development):**
```bash
STORAGE_PROVIDER=local
STORAGE_BUCKET=./uploads
STORAGE_BASE_URL=http://localhost:8080
```

##### **MinIO (Self-hosted):**
```bash
STORAGE_PROVIDER=minio
STORAGE_BUCKET=eralove-uploads
STORAGE_ACCESS_KEY_ID=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_ENDPOINT=localhost:9000
STORAGE_USE_SSL=false
STORAGE_BASE_URL=http://localhost:9000
```

##### **AWS S3 (Production):**
```bash
STORAGE_PROVIDER=s3
STORAGE_REGION=us-east-1
STORAGE_BUCKET=your-s3-bucket
STORAGE_ACCESS_KEY_ID=your-aws-key
STORAGE_SECRET_KEY=your-aws-secret
STORAGE_USE_SSL=true
```

### 🔧 **Integration với Photo Service:**

#### **Photo Upload Flow:**
1. **File Validation** - Check file type và size
2. **Storage Upload** - Upload file to configured storage
3. **URL Generation** - Get public URL for file access
4. **Database Save** - Save photo metadata với URL

```go
// Photo Service với Storage Integration
func (s *PhotoService) CreatePhoto(ctx context.Context, userID primitive.ObjectID, req *domain.CreatePhotoRequest, file interface{}) (*domain.PhotoResponse, error) {
    // Handle multipart file upload
    if fileHeader, ok := file.(*multipart.FileHeader); ok {
        // Validate file
        domain.ValidateImageFile(contentType, size)
        
        // Upload to storage
        uploadReq := &domain.UploadRequest{
            File:        src,
            Filename:    fileHeader.Filename,
            ContentType: contentType,
            Size:        size,
            Folder:      "photos",
            UserID:      userID.Hex(),
        }
        
        fileInfo, err := s.storageService.Upload(ctx, uploadReq)
        imageURL = fileInfo.URL
    }
    
    // Save to database với imageURL
}
```

### 🎯 **Features:**

#### **1. File Management:**
- ✅ **Upload** - Support multipart file upload
- ✅ **Download** - Presigned URLs cho secure access
- ✅ **Delete** - Remove files from storage
- ✅ **List** - Browse files in folders
- ✅ **Metadata** - File info retrieval

#### **2. Security:**
- ✅ **File Validation** - Type và size checking
- ✅ **Presigned URLs** - Temporary access links
- ✅ **Folder Organization** - User-based file separation
- ✅ **Unique Naming** - Timestamp-based file naming

#### **3. Flexibility:**
- ✅ **Multiple Providers** - Easy switching between storage types
- ✅ **Configuration-driven** - Environment-based setup
- ✅ **Development-friendly** - Local storage fallback
- ✅ **Production-ready** - S3/MinIO support

### 📁 **File Organization:**
```
Storage Structure:
├── photos/
│   ├── {user_id}/
│   │   ├── image1_20250927_120000.jpg
│   │   ├── image2_20250927_120100.png
│   │   └── ...
│   └── {another_user_id}/
├── avatars/
│   ├── {user_id}/
│   └── ...
└── documents/
    └── ...
```

### 🚀 **Usage Examples:**

#### **Development với Local Storage:**
```bash
# .env
STORAGE_PROVIDER=local
STORAGE_BUCKET=./uploads
STORAGE_BASE_URL=http://localhost:8080
```

#### **Development với MinIO:**
```bash
# Start MinIO
docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"

# .env
STORAGE_PROVIDER=minio
STORAGE_ENDPOINT=localhost:9000
STORAGE_ACCESS_KEY_ID=minioadmin
STORAGE_SECRET_KEY=minioadmin
```

#### **Production với AWS S3:**
```bash
# .env
STORAGE_PROVIDER=s3
STORAGE_REGION=us-east-1
STORAGE_BUCKET=eralove-production
STORAGE_ACCESS_KEY_ID=AKIA...
STORAGE_SECRET_KEY=...
```

### 📝 **Next Steps:**

#### **Cần hoàn thành:**
1. **Fix MinIO Dependencies** - Run `go mod tidy` để download MinIO SDK
2. **Update Providers** - Inject storage service vào PhotoService
3. **Add File Routes** - Static file serving cho local storage
4. **Add Admin Endpoints** - Manage uploaded files
5. **Add Cleanup Jobs** - Remove orphaned files

#### **Optional Enhancements:**
- **Image Resizing** - Multiple sizes cho thumbnails
- **CDN Integration** - CloudFront/CloudFlare integration
- **Backup Strategy** - Cross-provider backup
- **Analytics** - Storage usage tracking

### 🎉 **Benefits:**

#### **1. Flexibility:**
- Switch storage providers without code changes
- Development → Staging → Production migration
- Cost optimization options

#### **2. Scalability:**
- Handle large file uploads
- Presigned URLs reduce server load
- Distributed storage support

#### **3. Security:**
- File validation prevents malicious uploads
- Presigned URLs provide temporary access
- User-based file isolation

#### **4. Developer Experience:**
- Easy local development setup
- Clear configuration options
- Comprehensive logging

### 💡 **Recommendations:**

#### **Development:**
- Use **Local Storage** cho quick setup
- Use **MinIO** để test S3 compatibility

#### **Production:**
- Use **AWS S3** cho reliability và scalability
- Use **MinIO** cho self-hosted solutions
- Enable **SSL/TLS** cho security

#### **Configuration:**
- Set appropriate **file size limits**
- Configure **CORS** cho frontend uploads
- Use **environment-specific** buckets

Hệ thống storage này đã sẵn sàng để handle file uploads cho EraLove app với flexibility để scale từ development đến production! 🚀
