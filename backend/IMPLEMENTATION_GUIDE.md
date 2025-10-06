# Implementation Guide: Error Codes & Trace ID

## ✅ Đã hoàn thành

### 1. Error Code System
- File: `internal/domain/errors.go`
- Format: `HTTPCODE + 3 digits` (e.g., 400001, 409001, 500001)
- Đã định nghĩa 68 error codes

### 2. Response Structures
- File: `internal/handler/common.go`
- `ErrorResponse`: Có `code`, `error`, `message`, `trace_id`, `details`
- `SuccessResponse`: Có `success`, `data`, `message`, `trace_id`
- Helper: `getTraceID(c)` để lấy trace ID từ context

### 3. Updated Handlers
- ✅ `Register` handler - Đã có error codes và trace_id

## 📋 Cần cập nhật các handlers còn lại

### Pattern để update:

#### Error Response:
```go
return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
    Code:    int(domain.ErrCodeInvalidRequest),  // Thêm error code
    Error:   "Invalid request body",
    Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
    TraceID: getTraceID(c),  // Thêm trace ID
})
```

#### Success Response:
```go
return c.Status(fiber.StatusOK).JSON(SuccessResponse{
    Success: true,  // Thêm success flag
    Data:    data,
    Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "operation_successful", nil),
    TraceID: getTraceID(c),  // Thêm trace ID
})
```

### Handlers cần update:

#### user_handler.go
- [x] Register - ✅ Done
- [ ] Login
- [ ] GetProfile
- [ ] UpdateProfile
- [ ] DeleteAccount
- [ ] RefreshToken
- [ ] Logout
- [ ] VerifyEmail
- [ ] ResendVerificationEmail
- [ ] ForgotPassword
- [ ] ResetPassword

#### photo_handler.go (nếu có)
- [ ] CreatePhoto
- [ ] GetPhotos
- [ ] GetPhoto
- [ ] UpdatePhoto
- [ ] DeletePhoto

#### upload_handler.go
- [ ] UploadFile
- [ ] UploadMultipleFiles
- [ ] DeleteFile

## 🔧 Error Code Mapping

### Authentication Errors (401xxx)
- `401001` - Unauthorized
- `401002` - Invalid credentials
- `401003` - Invalid token
- `401004` - Token expired
- `401005` - Invalid verification token
- `401006` - Invalid reset token

### Validation Errors (400xxx)
- `400001` - Invalid request
- `400002` - Validation failed
- `400003` - Required field
- `400004` - Invalid format
- `400005` - Weak password
- `400006` - Password mismatch
- `400007` - Invalid email
- `400008` - Unsupported file type
- `400009` - File too large

### Conflict Errors (409xxx)
- `409001` - User already exists
- `409002` - Email already verified
- `409003` - Match request exists

### Not Found Errors (404xxx)
- `404001` - Resource not found
- `404002` - User not found
- `404003` - Photo not found
- `404004` - Event not found
- `404005` - Message not found
- `404006` - Match request not found
- `404007` - File not found
- `404008` - Conversation not found

### Server Errors (500xxx)
- `500001` - Internal error
- `500002` - Database error
- `500003` - Cache error
- `500004` - File upload failed
- `500005` - File delete failed

## 📝 Example Response

### Success Response:
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "message": "Registration successful! Please check your email to verify your account.",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error Response:
```json
{
  "code": 409001,
  "error": "Email already exists",
  "message": "User with this email already exists",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error with Details:
```json
{
  "code": 400002,
  "error": "Validation failed",
  "message": "Validation failed",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "details": {
    "email": "Email is required",
    "password": "Password must be at least 6 characters"
  }
}
```

## 🚀 Testing

```bash
# Test registration with trace ID
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-trace-123" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User",
    "date_of_birth": "1995-06-15",
    "gender": "male"
  }'

# Response will include trace_id
```

## 📌 Notes

- Trace ID được generate tự động bởi `requestid` middleware
- Có thể gửi custom trace ID qua header `X-Request-ID`
- Tất cả responses (success/error) đều có trace_id
- Error codes giúp client xử lý lỗi chính xác hơn
- I18n messages được load từ `messages/*.json`
