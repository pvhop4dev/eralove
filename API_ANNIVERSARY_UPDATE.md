# Update Anniversary Date API

## Endpoint
```
PUT /api/v1/users/profile
```

## Authentication
Requires Bearer token in Authorization header.

## Request Body
```json
{
  "anniversary_date": "2024-01-15"
}
```

Có thể kết hợp với các fields khác:
```json
{
  "name": "John Doe",
  "anniversary_date": "2024-01-15",
  "partner_name": "Jane Doe"
}
```

## Response
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "John Doe",
  "email": "john@example.com",
  "partner_id": "507f191e810c19729de860ea",
  "partner_name": "Jane Doe",
  "match_code": "507f191e810c19729de860ea_507f1f77bcf86cd799439011",
  "matched_at": "2024-01-01T00:00:00Z",
  "anniversary_date": "2024-01-15T00:00:00Z",
  "created_at": "2023-12-01T00:00:00Z",
  "updated_at": "2024-01-20T10:30:00Z"
}
```

## Business Logic

### 1. Validation
- ✅ User phải đã matched (có `match_code`)
- ❌ Nếu chưa match → Error: "cannot set anniversary date: user is not matched"

### 2. Auto-sync với Partner
Khi user A update anniversary date:
1. Update anniversary date của user A
2. **Tự động update** anniversary date của partner (user B)
3. Cả 2 người luôn có cùng anniversary date

### 3. Use Cases

#### Use Case 1: Set Anniversary Date lần đầu
```javascript
// User vừa match nhưng chưa có anniversary date
PUT /api/v1/users/profile
{
  "anniversary_date": "2024-02-14"
}

// Result: Cả 2 users đều có anniversary_date = "2024-02-14"
```

#### Use Case 2: Update Anniversary Date
```javascript
// User muốn sửa lại anniversary date
PUT /api/v1/users/profile
{
  "anniversary_date": "2024-03-01"
}

// Result: Cả 2 users đều có anniversary_date = "2024-03-01"
```

#### Use Case 3: Error - Chưa match
```javascript
// User chưa match với ai
PUT /api/v1/users/profile
{
  "anniversary_date": "2024-02-14"
}

// Error Response (400):
{
  "error": "Failed to update profile",
  "message": "cannot set anniversary date: user is not matched"
}
```

## Frontend Implementation

### React Example
```typescript
// services/api.ts
export const updateProfile = (data: {
  name?: string;
  anniversary_date?: string;
  partner_name?: string;
}) => api.put('/users/profile', data);

// pages/Settings.tsx
import { useState } from 'react';
import { updateProfile } from '../services/api';

function Settings() {
  const [anniversaryDate, setAnniversaryDate] = useState('');
  const [loading, setLoading] = useState(false);

  const handleUpdateAnniversary = async () => {
    setLoading(true);
    try {
      await updateProfile({
        anniversary_date: anniversaryDate
      });
      toast.success('Anniversary date updated successfully!');
      // Refresh user data
      queryClient.invalidateQueries(['user']);
    } catch (error) {
      if (error.response?.data?.message?.includes('not matched')) {
        toast.error('You need to be matched with someone first');
      } else {
        toast.error('Failed to update anniversary date');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="settings-page">
      <h2>Anniversary Date</h2>
      <input
        type="date"
        value={anniversaryDate}
        onChange={(e) => setAnniversaryDate(e.target.value)}
        className="form-control"
      />
      <button 
        onClick={handleUpdateAnniversary}
        disabled={loading || !anniversaryDate}
        className="btn btn-primary"
      >
        {loading ? 'Updating...' : 'Update Anniversary Date'}
      </button>
    </div>
  );
}
```

### Vue Example
```vue
<template>
  <div class="settings-page">
    <h2>Anniversary Date</h2>
    <input
      type="date"
      v-model="anniversaryDate"
      class="form-control"
    />
    <button 
      @click="handleUpdateAnniversary"
      :disabled="loading || !anniversaryDate"
      class="btn btn-primary"
    >
      {{ loading ? 'Updating...' : 'Update Anniversary Date' }}
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { updateProfile } from '@/services/api';
import { useToast } from '@/composables/useToast';

const anniversaryDate = ref('');
const loading = ref(false);
const toast = useToast();

const handleUpdateAnniversary = async () => {
  loading.value = true;
  try {
    await updateProfile({
      anniversary_date: anniversaryDate.value
    });
    toast.success('Anniversary date updated successfully!');
  } catch (error) {
    if (error.response?.data?.message?.includes('not matched')) {
      toast.error('You need to be matched with someone first');
    } else {
      toast.error('Failed to update anniversary date');
    }
  } finally {
    loading.value = false;
  }
};
</script>
```

## Testing

### Manual Testing
```bash
# 1. Login to get token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# 2. Update anniversary date
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "anniversary_date": "2024-02-14"
  }'

# 3. Verify both users have same date
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Automated Testing
```go
func TestUpdateAnniversaryDate(t *testing.T) {
    // Test case 1: Update anniversary date successfully
    // Test case 2: Error when user not matched
    // Test case 3: Partner's date is also updated
}
```

## Notes

### Important Points
1. **Sync tự động**: Không cần gọi API 2 lần, partner tự động được update
2. **Validation**: Chỉ matched users mới có thể set anniversary date
3. **Idempotent**: Có thể update nhiều lần, giá trị mới nhất sẽ được lưu
4. **Atomic**: Nếu update partner fail, current user vẫn được update

### Future Enhancements
- [ ] Add notification khi partner update anniversary date
- [ ] Add history log cho anniversary date changes
- [ ] Add permission: chỉ cho phép update 1 lần/tháng
- [ ] Add confirmation từ partner trước khi apply change
