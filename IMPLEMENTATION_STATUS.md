# Match System Implementation Status

## ✅ Completed

### 1. Domain Models (100%)
- ✅ `user.go` - Added `MatchCode` and `MatchedAt` fields
- ✅ `event.go` - Replaced `UserID`/`PartnerID` with `MatchCode`
- ✅ `photo.go` - Replaced `UserID`/`PartnerID` with `MatchCode`
- ✅ `match_utils.go` - Created `GenerateMatchCode()` utility function
- ✅ `MATCH_SYSTEM_DESIGN.md` - Complete system documentation

### 2. Repositories (90%)
- ✅ `event_repository.go` - Updated all methods to use `MatchCode`
  - `GetByMatchCode()`
  - `GetByMatchCodeAndDateRange()`
  - `GetByMatchCodeAndDate()`
  - `GetUpcomingByMatchCode()`
  - `DeleteByMatchCode()`
- ✅ `photo_repository_new.go` - New implementation with `MatchCode`
  - `GetByMatchCode()`
  - `GetByMatchCodeAndDate()`
  - `DeleteByMatchCode()`
  - `SearchByMatchCode()`

### 3. Services (70%)
- ✅ `event_service.go` - Updated to use `MatchCode`
  - Added `userRepo` dependency
  - `CreateEvent()` - Gets user's match code
  - `GetCoupleEvents()` - Replaces `GetUserEvents()`
  - `GetEvent()`, `UpdateEvent()`, `DeleteEvent()` - Verify match code
- ✅ `providers.go` - Updated `ProvideEventService()` signature

### 4. Handlers (50%)
- ✅ `event_handler.go` - Updated `GetEvents()` to call `GetCoupleEvents()`

## 🔧 Remaining Work

### Backend

#### 1. Replace Old PhotoRepository
```bash
# Delete old file
rm backend/internal/repository/photo_repository.go

# Rename new file
mv backend/internal/repository/photo_repository_new.go backend/internal/repository/photo_repository.go
```

#### 2. Update PhotoService
File: `backend/internal/service/photo_service.go`

Changes needed:
- Add `userRepo` dependency
- Update `CreatePhoto()` to get user's match code
- Update `GetCouplePhotos()` to replace `GetUserPhotos()`
- Update authorization checks to use match code

#### 3. Update UserService
File: `backend/internal/service/user_service.go`

Add `UnmatchPartner()` method:
```go
func (s *UserService) UnmatchPartner(ctx context.Context, userID primitive.ObjectID) error {
    // 1. Get user
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }
    
    if user.MatchCode == "" {
        return fmt.Errorf("user is not matched")
    }
    
    // 2. Delete all events with match code
    if err := s.eventRepo.DeleteByMatchCode(user.MatchCode); err != nil {
        return err
    }
    
    // 3. Delete all photos with match code
    if err := s.photoRepo.DeleteByMatchCode(ctx, user.MatchCode); err != nil {
        return err
    }
    
    // 4. Get partner
    if user.PartnerID != nil {
        partner, _ := s.userRepo.GetByID(ctx, *user.PartnerID)
        if partner != nil {
            // Clear partner's match fields
            partner.PartnerID = nil
            partner.MatchCode = ""
            partner.MatchedAt = nil
            partner.AnniversaryDate = nil
            s.userRepo.Update(ctx, partner.ID, partner)
        }
    }
    
    // 5. Clear user's match fields
    user.PartnerID = nil
    user.MatchCode = ""
    user.MatchedAt = nil
    user.AnniversaryDate = nil
    
    return s.userRepo.Update(ctx, userID, user)
}
```

#### 4. Update MatchRequestService
File: `backend/internal/service/match_request_service.go`

Update `RespondToMatchRequest()` when action is "accept":
```go
if action == "accept" {
    // Generate match code
    matchCode := domain.GenerateMatchCode(sender.ID, receiver.ID)
    now := time.Now()
    
    // Update sender
    sender.PartnerID = &receiver.ID
    sender.MatchCode = matchCode
    sender.MatchedAt = &now
    sender.AnniversaryDate = &matchRequest.AnniversaryDate
    s.userRepo.Update(ctx, sender.ID, sender)
    
    // Update receiver
    receiver.PartnerID = &sender.ID
    receiver.MatchCode = matchCode
    receiver.MatchedAt = &now
    receiver.AnniversaryDate = &matchRequest.AnniversaryDate
    s.userRepo.Update(ctx, receiver.ID, receiver)
}
```

#### 5. Update PhotoHandler
File: `backend/internal/handler/photo_handler.go`

Update `GetPhotos()` to call `GetCouplePhotos()` instead of `GetUserPhotos()`

#### 6. Add Unmatch Endpoint
File: `backend/internal/handler/user_handler.go`

Add new endpoint:
```go
// UnmatchPartner godoc
// @Summary Unmatch from partner
// @Description Break match with partner and delete all shared data
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/unmatch [post]
func (h *UserHandler) UnmatchPartner(c *fiber.Ctx) error {
    userID := getUserIDFromContext(c)
    
    if err := h.userService.UnmatchPartner(c.Context(), userID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Error: "Failed to unmatch",
            Message: err.Error(),
        })
    }
    
    return c.JSON(SuccessResponse{
        Message: "Successfully unmatched from partner",
    })
}
```

Register route in `backend/internal/app/routes.go`:
```go
users.Post("/unmatch", userHandler.UnmatchPartner)
```

#### 7. Update Wire Dependency Injection
File: `backend/internal/app/wire.go`

Update providers to include new dependencies.

### Frontend

#### 1. Update API Types
File: `frontend/src/types/api.ts`

Update interfaces:
```typescript
export interface User {
  id: string;
  name: string;
  email: string;
  partner_id?: string;
  partner_name?: string;
  match_code?: string;  // NEW
  matched_at?: string;  // NEW
  anniversary_date?: string;
  // ... other fields
}

export interface Event {
  id: string;
  match_code: string;  // CHANGED from user_id
  title: string;
  date: string;
  // ... other fields
}

export interface Photo {
  id: string;
  match_code: string;  // CHANGED from user_id
  title: string;
  image_url: string;
  // ... other fields
}
```

#### 2. Update API Calls
File: `frontend/src/services/api.ts`

Remove `partner_id` parameter from event and photo queries:
```typescript
// Before
export const getEvents = (params: { 
  page?: number; 
  limit?: number; 
  year?: number; 
  month?: number;
  partner_id?: string;  // REMOVE THIS
}) => api.get('/events', { params });

// After
export const getEvents = (params: { 
  page?: number; 
  limit?: number; 
  year?: number; 
  month?: number;
}) => api.get('/events', { params });
```

#### 3. Add Unmatch Functionality
File: `frontend/src/services/api.ts`

```typescript
export const unmatchPartner = () => api.post('/users/unmatch');
```

File: `frontend/src/pages/Settings.tsx` or similar

Add unmatch button:
```typescript
const handleUnmatch = async () => {
  if (confirm('Are you sure? This will delete all shared events and photos.')) {
    try {
      await unmatchPartner();
      // Refresh user data
      // Show success message
      // Redirect to home
    } catch (error) {
      // Show error message
    }
  }
};
```

#### 4. Update Components
Files to check:
- `frontend/src/pages/Calendar.tsx` - Remove partner_id from API calls
- `frontend/src/pages/Photos.tsx` - Remove partner_id from API calls
- `frontend/src/components/EventForm.tsx` - No changes needed (backend handles match_code)
- `frontend/src/components/PhotoUpload.tsx` - No changes needed (backend handles match_code)

## Database Migration

### Create Indexes
```javascript
// MongoDB shell
use eralove

// Users collection
db.users.createIndex({ "match_code": 1 })
db.users.createIndex({ "partner_id": 1 })

// Events collection
db.events.createIndex({ "match_code": 1, "date": -1 })
db.events.createIndex({ "match_code": 1, "event_type": 1 })

// Photos collection
db.photos.createIndex({ "match_code": 1, "date": -1 })
db.photos.createIndex({ "match_code": 1, "created_at": -1 })
```

### Migrate Existing Data (if any)
```javascript
// For each matched user pair, generate match code and update records
db.users.find({ partner_id: { $exists: true, $ne: null } }).forEach(user => {
    const partner = db.users.findOne({ _id: user.partner_id });
    if (partner) {
        // Generate match code (implement the same logic as Go function)
        const ids = [user._id.str, partner._id.str].sort();
        const matchCode = generateMatchCodeInJS(ids[0], ids[1]);
        
        // Update users
        db.users.updateOne({ _id: user._id }, { $set: { match_code: matchCode } });
        db.users.updateOne({ _id: partner._id }, { $set: { match_code: matchCode } });
        
        // Update events
        db.events.updateMany(
            { 
                $or: [
                    { user_id: user._id, partner_id: partner._id },
                    { user_id: partner._id, partner_id: user._id }
                ]
            },
            { 
                $set: { match_code: matchCode },
                $unset: { user_id: "", partner_id: "" }
            }
        );
        
        // Update photos
        db.photos.updateMany(
            { 
                $or: [
                    { user_id: user._id, partner_id: partner._id },
                    { user_id: partner._id, partner_id: user._id }
                ]
            },
            { 
                $set: { match_code: matchCode },
                $unset: { user_id: "", partner_id: "" }
            }
        );
    }
});
```

## Testing Checklist

### Backend
- [ ] Test match request accept generates correct match code
- [ ] Test event creation requires match code
- [ ] Test photo creation requires match code
- [ ] Test GetCoupleEvents returns correct events
- [ ] Test GetCouplePhotos returns correct photos
- [ ] Test unmatch deletes all shared data
- [ ] Test authorization checks use match code

### Frontend
- [ ] Test calendar displays couple events
- [ ] Test photo gallery displays couple photos
- [ ] Test unmatch button works
- [ ] Test UI updates after unmatch
- [ ] Test match request flow

## Priority Order

1. **High Priority** (Core functionality)
   - Replace PhotoRepository file
   - Update PhotoService
   - Update UserService with UnmatchPartner
   - Update MatchRequestService to generate match code

2. **Medium Priority** (Features)
   - Add unmatch endpoint and handler
   - Update frontend API types
   - Update frontend API calls
   - Add unmatch UI

3. **Low Priority** (Polish)
   - Database migration script
   - Comprehensive testing
   - Documentation updates

## Estimated Time
- Backend completion: 2-3 hours
- Frontend updates: 1-2 hours
- Testing: 1 hour
- **Total: 4-6 hours**
