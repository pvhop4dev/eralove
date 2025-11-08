# Anniversary Date - Vấn đề và Giải pháp

## 🐛 Vấn đề

### 1. **Type Mismatch**
```go
// MatchRequest model
type MatchRequest struct {
    AnniversaryDate time.Time  // NOT a pointer
}

// User model  
type User struct {
    AnniversaryDate *time.Time  // IS a pointer
}
```

Khi accept match request, code cũ:
```go
sender.AnniversaryDate = &matchRequest.AnniversaryDate  // ✅ OK
```

Nhưng nếu `matchRequest.AnniversaryDate` là zero value (`0001-01-01`), sẽ lưu ngày sai!

### 2. **Không cho phép Receiver override**
- Sender gửi request với anniversary date
- Receiver accept nhưng không thể thay đổi date
- Cả 2 bị stuck với date mà sender chọn

### 3. **Thiếu PartnerName**
- Khi match, không set `PartnerName` cho cả 2 users
- Frontend phải query thêm để lấy partner name

## ✅ Giải pháp

### 1. **Allow Receiver Override**

#### Request Model
```go
type RespondToMatchRequestRequest struct {
    Action          string     `json:"action" validate:"required,oneof=accept reject"`
    AnniversaryDate *time.Time `json:"anniversary_date,omitempty"` // Optional override
}
```

#### Logic Priority
```go
var finalAnniversaryDate time.Time

if req.AnniversaryDate != nil {
    // Priority 1: Receiver's choice (when accepting)
    finalAnniversaryDate = *req.AnniversaryDate
} else {
    // Priority 2: Sender's original date
    finalAnniversaryDate = matchRequest.AnniversaryDate
}
```

### 2. **Auto-set PartnerName**
```go
// Update sender
sender.PartnerName = receiver.Name

// Update receiver
receiver.PartnerName = sender.Name
```

### 3. **Better Logging**
```go
s.logger.Info("Match created successfully",
    zap.String("match_code", matchCode),
    zap.String("sender_id", sender.ID.Hex()),
    zap.String("receiver_id", receiver.ID.Hex()),
    zap.Time("anniversary_date", finalAnniversaryDate))
```

## 📝 API Usage

### Scenario 1: Accept với date gốc
```bash
POST /api/v1/match-requests/{id}/respond
{
  "action": "accept"
}
# Sử dụng anniversary_date từ match request gốc
```

### Scenario 2: Accept và thay đổi date
```bash
POST /api/v1/match-requests/{id}/respond
{
  "action": "accept",
  "anniversary_date": "2024-02-14T00:00:00Z"
}
# Override với date mới
```

### Scenario 3: Reject
```bash
POST /api/v1/match-requests/{id}/respond
{
  "action": "reject"
}
# anniversary_date không quan trọng
```

## 🔄 Flow hoàn chỉnh

### 1. Sender gửi request
```javascript
POST /api/v1/match-requests
{
  "receiver_email": "partner@example.com",
  "anniversary_date": "2024-01-01",
  "message": "Let's be together!"
}
```

### 2. Receiver nhận được notification
```javascript
GET /api/v1/match-requests/received
// Response:
[{
  "id": "...",
  "sender_id": "...",
  "anniversary_date": "2024-01-01",  // Date mà sender đề xuất
  "message": "Let's be together!"
}]
```

### 3. Receiver có 2 lựa chọn

#### Option A: Đồng ý với date gốc
```javascript
POST /api/v1/match-requests/{id}/respond
{
  "action": "accept"
}
// Result: Cả 2 có anniversary_date = "2024-01-01"
```

#### Option B: Thay đổi date
```javascript
POST /api/v1/match-requests/{id}/respond
{
  "action": "accept",
  "anniversary_date": "2024-02-14"  // Valentine's Day!
}
// Result: Cả 2 có anniversary_date = "2024-02-14"
```

### 4. Sau khi match, có thể update
```javascript
PUT /api/v1/users/profile
{
  "anniversary_date": "2024-03-01"
}
// Cả 2 users đều được update
```

## 🎯 Benefits

### 1. **Flexibility**
- Receiver có quyền quyết định final date
- Không bị stuck với date mà sender chọn

### 2. **User Experience**
- Có thể discuss và agree trên date
- Frontend có thể show date suggestion từ sender
- Receiver có thể accept hoặc propose new date

### 3. **Data Consistency**
- Cả 2 users luôn có cùng anniversary date
- PartnerName được set tự động
- Không có zero value dates

## 🧪 Testing

### Test Case 1: Accept với date gốc
```go
func TestAcceptWithOriginalDate(t *testing.T) {
    // Send request with date "2024-01-01"
    // Accept without override
    // Assert: Both users have "2024-01-01"
}
```

### Test Case 2: Accept với override
```go
func TestAcceptWithOverride(t *testing.T) {
    // Send request with date "2024-01-01"
    // Accept with override "2024-02-14"
    // Assert: Both users have "2024-02-14"
}
```

### Test Case 3: Update sau khi match
```go
func TestUpdateAfterMatch(t *testing.T) {
    // Match with date "2024-01-01"
    // User A updates to "2024-03-01"
    // Assert: Both users have "2024-03-01"
}
```

## 📱 Frontend Implementation

### Accept Match Request Form
```typescript
interface AcceptMatchFormProps {
  matchRequest: MatchRequest;
  onAccept: (data: RespondRequest) => void;
}

function AcceptMatchForm({ matchRequest, onAccept }: AcceptMatchFormProps) {
  const [useCustomDate, setUseCustomDate] = useState(false);
  const [anniversaryDate, setAnniversaryDate] = useState('');
  
  const handleAccept = () => {
    const data: RespondRequest = {
      action: 'accept'
    };
    
    // Only include anniversary_date if user wants to override
    if (useCustomDate && anniversaryDate) {
      data.anniversary_date = anniversaryDate;
    }
    
    onAccept(data);
  };
  
  return (
    <div className="accept-form">
      <h3>Accept Match Request</h3>
      
      <div className="date-section">
        <p>Suggested Anniversary Date: {matchRequest.anniversary_date}</p>
        
        <label>
          <input
            type="checkbox"
            checked={useCustomDate}
            onChange={(e) => setUseCustomDate(e.target.checked)}
          />
          Use different date
        </label>
        
        {useCustomDate && (
          <input
            type="date"
            value={anniversaryDate}
            onChange={(e) => setAnniversaryDate(e.target.value)}
          />
        )}
      </div>
      
      <button onClick={handleAccept}>Accept</button>
    </div>
  );
}
```

### API Service
```typescript
export const respondToMatchRequest = (
  requestId: string,
  data: {
    action: 'accept' | 'reject';
    anniversary_date?: string;
  }
) => api.post(`/match-requests/${requestId}/respond`, data);
```

## 🔍 Debug Tips

### Check anniversary date in DB
```javascript
db.users.find({ match_code: { $exists: true } }).forEach(user => {
    print(`User: ${user.name}`);
    print(`Anniversary: ${user.anniversary_date}`);
    print(`Partner: ${user.partner_name}`);
    print('---');
});
```

### Verify both users have same date
```javascript
db.users.aggregate([
    { $match: { match_code: { $exists: true, $ne: "" } } },
    { $group: {
        _id: "$match_code",
        users: { $push: { name: "$name", date: "$anniversary_date" } }
    }},
    { $match: { "users.1": { $exists: true } } }
]);
```

## 📊 Summary

| Feature | Before | After |
|---------|--------|-------|
| Receiver can override date | ❌ No | ✅ Yes |
| PartnerName auto-set | ❌ No | ✅ Yes |
| Zero value dates | ⚠️ Possible | ✅ Prevented |
| Update after match | ❌ No | ✅ Yes (both users) |
| Logging | ⚠️ Basic | ✅ Detailed |

## ✅ Files Changed

1. `backend/internal/domain/match_request.go` - Updated interface
2. `backend/internal/service/match_request_service.go` - New logic
3. `backend/internal/handler/match_request_handler.go` - Pass request object
4. `backend/internal/domain/user.go` - Added anniversary_date to UpdateUserRequest
5. `backend/internal/service/user_service.go` - Update logic with partner sync

All changes are backward compatible! 🎉
