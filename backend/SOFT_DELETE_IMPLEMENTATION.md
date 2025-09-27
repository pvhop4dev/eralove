# Soft Delete Implementation Guide

## Overview
Đã implement soft delete pattern cho tất cả các models trong EraLove backend để bảo toàn dữ liệu và cho phép khôi phục.

## ✅ Models đã được cập nhật với `deleted_at` field:

### 1. **User Model** (`internal/domain/user.go`)
```go
type User struct {
    // ... other fields
    DeletedAt *time.Time `json:"-" bson:"deleted_at,omitempty"`
}
```

### 2. **Photo Model** (`internal/domain/photo.go`)
```go
type Photo struct {
    // ... other fields
    DeletedAt *time.Time `json:"-" bson:"deleted_at,omitempty"`
}
```

### 3. **Event Model** (`internal/domain/event.go`)
```go
type Event struct {
    // ... other fields
    DeletedAt *time.Time `json:"-" bson:"deleted_at,omitempty"`
}
```

### 4. **Message Model** (`internal/domain/message.go`)
```go
type Message struct {
    // ... other fields
    DeletedAt *time.Time `json:"-" bson:"deleted_at,omitempty"`
}
```

## ✅ Repository Pattern đã được cập nhật:

### **User Repository** (`internal/repository/user_repository.go`)

#### **Query Methods (chỉ lấy active records):**
- `GetByID()` - Chỉ lấy user chưa bị xóa
- `GetByEmail()` - Chỉ lấy user active
- `GetByEmailVerificationToken()` - Chỉ lấy user active
- `GetByPasswordResetToken()` - Chỉ lấy user active
- `List()` - Chỉ list user active
- `Update()` - Chỉ update user active

#### **Soft Delete Methods:**
```go
// Soft delete - set deleted_at timestamp
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error

// Restore soft deleted user
func (r *UserRepository) Restore(ctx context.Context, id primitive.ObjectID) error

// Permanently delete from database
func (r *UserRepository) HardDelete(ctx context.Context, id primitive.ObjectID) error

// List deleted users
func (r *UserRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*domain.User, error)
```

#### **Filter Helpers:**
```go
// Active user filter
func getActiveUserFilter() bson.M {
    return bson.M{
        "is_active": true,
        "deleted_at": bson.M{"$exists": false},
    }
}

// Active user filter with conditions
func getActiveUserFilterWithCondition(condition bson.M) bson.M
```

## ✅ Soft Delete Helper Utility (`internal/repository/soft_delete_helper.go`)

### **General Purpose Filters:**
```go
// For any model
SoftDelete.GetActiveFilter()                           // deleted_at not exists
SoftDelete.GetActiveFilterWithCondition(condition)    // active + custom conditions
SoftDelete.GetDeletedFilter()                         // deleted_at exists
SoftDelete.GetDeletedFilterWithCondition(condition)   // deleted + custom conditions
```

### **Specific Filters:**
```go
// By ID
SoftDelete.GetActiveFilterByID(id)

// By User ID  
SoftDelete.GetActiveFilterByUserID(userID)

// By Couple (for photos, events)
SoftDelete.GetActiveFilterByCoupleID(userID, partnerID)

// By Conversation (for messages)
SoftDelete.GetActiveFilterByConversation(userID, partnerID)
```

### **Update Operations:**
```go
// Soft delete update
SoftDelete.CreateSoftDeleteUpdate()

// Restore update
SoftDelete.CreateRestoreUpdate()

// Soft delete with additional fields
SoftDelete.CreateSoftDeleteUpdateWithFields(additionalFields)
```

## 🔧 Implementation Examples:

### **Photo Repository Example:**
```go
// Get active photos only
func (r *PhotoRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Photo, error) {
    filter := SoftDelete.GetActiveFilterByUserID(userID)
    cursor, err := r.collection.Find(ctx, filter)
    // ... rest of implementation
}

// Soft delete photo
func (r *PhotoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    filter := SoftDelete.GetActiveFilterByID(id)
    update := SoftDelete.CreateSoftDeleteUpdate()
    
    result, err := r.collection.UpdateOne(ctx, filter, update)
    // ... error handling
}
```

### **Event Repository Example:**
```go
// Get couple's events (active only)
func (r *EventRepository) GetByCoupleID(ctx context.Context, userID, partnerID primitive.ObjectID) ([]*domain.Event, error) {
    filter := SoftDelete.GetActiveFilterByCoupleID(userID, partnerID)
    cursor, err := r.collection.Find(ctx, filter)
    // ... rest of implementation
}
```

### **Message Repository Example:**
```go
// Get conversation messages (active only)
func (r *MessageRepository) FindConversation(ctx context.Context, userID, partnerID primitive.ObjectID) ([]*domain.Message, error) {
    filter := SoftDelete.GetActiveFilterByConversation(userID, partnerID)
    cursor, err := r.collection.Find(ctx, filter)
    // ... rest of implementation
}
```

## 📋 TODO - Cần implement cho các repositories khác:

### **Photo Repository:**
- [ ] Cập nhật tất cả query methods để sử dụng soft delete filters
- [ ] Implement `Restore()` và `HardDelete()` methods
- [ ] Implement `ListDeleted()` method

### **Event Repository:**
- [ ] Cập nhật tất cả query methods để sử dụng soft delete filters  
- [ ] Implement `Restore()` và `HardDelete()` methods
- [ ] Implement `ListDeleted()` method

### **Message Repository:**
- [ ] Cập nhật tất cả query methods để sử dụng soft delete filters
- [ ] Implement `Restore()` và `HardDelete()` methods
- [ ] Implement `ListDeleted()` method

## 🔒 Security & Best Practices:

### **1. Field Visibility:**
```go
DeletedAt *time.Time `json:"-" bson:"deleted_at,omitempty"`
```
- `json:"-"` - Không expose trong API response
- `bson:"deleted_at,omitempty"` - MongoDB field mapping

### **2. Query Safety:**
- Tất cả queries mặc định chỉ lấy active records
- Phải explicitly query để lấy deleted records
- Sử dụng indexes cho performance

### **3. Logging:**
```go
r.logger.Info("User soft deleted successfully", 
    zap.String("user_id", id.Hex()),
    zap.Time("deleted_at", now))

r.logger.Warn("User hard deleted permanently", 
    zap.String("user_id", id.Hex()))
```

### **4. Database Indexes:**
```javascript
// MongoDB indexes cần tạo
db.users.createIndex({ "deleted_at": 1 })
db.photos.createIndex({ "deleted_at": 1 })
db.events.createIndex({ "deleted_at": 1 })
db.messages.createIndex({ "deleted_at": 1 })

// Compound indexes
db.photos.createIndex({ "user_id": 1, "deleted_at": 1 })
db.events.createIndex({ "user_id": 1, "deleted_at": 1 })
db.messages.createIndex({ "sender_id": 1, "receiver_id": 1, "deleted_at": 1 })
```

## 🎯 Benefits:

### **1. Data Protection:**
- Không mất dữ liệu khi user "xóa"
- Có thể khôi phục nếu cần
- Audit trail cho compliance

### **2. Performance:**
- Queries chỉ scan active records
- Indexes optimize performance
- Soft delete nhanh hơn hard delete

### **3. Business Logic:**
- User có thể "undo" delete
- Admin có thể restore data
- Analytics vẫn có historical data

### **4. Compliance:**
- GDPR: Có thể hard delete khi cần
- Data retention policies
- Legal requirements

## 🚀 Next Steps:

1. **Implement remaining repositories** (Photo, Event, Message)
2. **Add database indexes** cho performance
3. **Create admin endpoints** để manage deleted records
4. **Add cleanup jobs** để hard delete old records
5. **Update API documentation** với soft delete behavior
6. **Add unit tests** cho soft delete functionality

## 📝 Usage Notes:

- **Default behavior**: Tất cả queries chỉ lấy active records
- **Explicit deletion**: Phải explicitly query deleted records
- **Restore capability**: Có thể restore bất kỳ lúc nào
- **Hard delete**: Chỉ dùng khi thực sự cần thiết (GDPR, cleanup)
- **Performance**: Sử dụng indexes để optimize queries
