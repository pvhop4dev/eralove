# Match System Design

## Overview
The new match system uses a **shared match code** approach where matched couples share a unique identifier for all their events and photos. When users unmatch, all shared data is automatically cleaned up.

## Key Concepts

### 1. Match Code Generation
- **Deterministic**: Same two users always generate the same code
- **Algorithm**: 
  1. Sort both user IDs alphabetically
  2. Concatenate sorted IDs
  3. Generate SHA256 hash
  4. Take first 16 characters as match code
- **Example**: `userA + userB → "a3f5c8d9e1b2f4a6"`

### 2. Data Model Changes

#### User Model
```go
type User struct {
    // ... existing fields
    PartnerID       *primitive.ObjectID  // Reference to partner
    PartnerName     string               // Partner's name
    MatchCode       string               // Unique code for the couple
    MatchedAt       *time.Time           // When they matched
    AnniversaryDate *time.Time           // Their anniversary
}
```

#### Event Model
```go
type Event struct {
    ID          primitive.ObjectID
    MatchCode   string      // Replaces UserID + PartnerID
    Title       string
    Date        time.Time
    // ... other fields
}
```

#### Photo Model
```go
type Photo struct {
    ID          primitive.ObjectID
    MatchCode   string      // Replaces UserID + PartnerID
    Title       string
    ImageURL    string
    Date        time.Time
    // ... other fields
}
```

## Workflow

### Match Request Flow
1. **User A sends match request** to User B
   - Includes anniversary date and optional message
   - Status: `pending`

2. **User B accepts request**
   - Generate match code: `GenerateMatchCode(userA_ID, userB_ID)`
   - Update both users:
     - Set `PartnerID` to each other
     - Set `MatchCode` to generated code
     - Set `MatchedAt` to current time
     - Set `AnniversaryDate` from request
   - Match request status: `accepted`

3. **Create/Access shared data**
   - All events use `MatchCode` instead of individual user IDs
   - All photos use `MatchCode` instead of individual user IDs
   - Both users can access all shared events/photos via their `MatchCode`

### Unmatch Flow
1. **User initiates unmatch**
   - Call `UnmatchPartner(userID)`

2. **System cleanup**
   - Get user's `MatchCode`
   - Delete all events with that `MatchCode`
   - Delete all photos with that `MatchCode`
   - Clear both users' match fields:
     - `PartnerID` → `nil`
     - `MatchCode` → `""`
     - `MatchedAt` → `nil`
     - `AnniversaryDate` → `nil`

## API Changes

### Repository Methods

#### EventRepository
```go
// Old methods (removed)
GetByUserID(userID, limit, offset)
GetByCoupleID(userID, partnerID, limit, offset)

// New methods
GetByMatchCode(matchCode, limit, offset)
GetByMatchCodeAndDateRange(matchCode, startDate, endDate)
DeleteByMatchCode(matchCode)  // For unmatch cleanup
```

#### PhotoRepository
```go
// Old methods (removed)
GetByUserID(userID, limit, offset)
GetByCoupleID(userID, partnerID, limit, offset)

// New methods
GetByMatchCode(matchCode, limit, offset)
GetByMatchCodeAndDate(matchCode, date)
DeleteByMatchCode(matchCode)  // For unmatch cleanup
```

### Service Methods

#### UserService
```go
// New method
UnmatchPartner(ctx, userID) error
```

#### EventService
```go
// Updated method
GetCoupleEvents(ctx, userID, year, month, page, limit)
// Internally: Get user → Get MatchCode → Query by MatchCode
```

#### PhotoService
```go
// Updated method
GetCouplePhotos(ctx, userID, page, limit)
// Internally: Get user → Get MatchCode → Query by MatchCode
```

## Benefits

### 1. Data Isolation
- Each couple's data is completely isolated by match code
- No mixing of data between different relationships

### 2. Automatic Cleanup
- When unmatch happens, all shared data is removed
- No orphaned events or photos

### 3. Simplified Queries
- Single field query instead of complex OR conditions
- Better database performance with proper indexing

### 4. Consistency
- Same match code regardless of who queries
- No need to check both (userA, userB) and (userB, userA)

### 5. Privacy
- When relationship ends, all shared memories are removed
- Clean slate for new relationships

## Database Indexes

### Users Collection
```javascript
db.users.createIndex({ "match_code": 1 })
db.users.createIndex({ "partner_id": 1 })
```

### Events Collection
```javascript
db.events.createIndex({ "match_code": 1, "date": -1 })
db.events.createIndex({ "match_code": 1, "event_type": 1 })
```

### Photos Collection
```javascript
db.photos.createIndex({ "match_code": 1, "date": -1 })
db.photos.createIndex({ "match_code": 1, "created_at": -1 })
```

## Migration Notes

### For Existing Data
If you have existing data with `user_id` and `partner_id`:

1. Generate match codes for all matched couples
2. Update events and photos with new match codes
3. Remove old `user_id` and `partner_id` fields
4. Update indexes

### Migration Script Example
```javascript
// For each matched user pair
db.users.find({ partner_id: { $exists: true, $ne: null } }).forEach(user => {
    const partner = db.users.findOne({ _id: user.partner_id });
    if (partner) {
        const matchCode = generateMatchCode(user._id, partner._id);
        
        // Update users
        db.users.updateOne({ _id: user._id }, { $set: { match_code: matchCode } });
        
        // Update events
        db.events.updateMany(
            { user_id: user._id, partner_id: partner._id },
            { $set: { match_code: matchCode }, $unset: { user_id: "", partner_id: "" } }
        );
        
        // Update photos
        db.photos.updateMany(
            { user_id: user._id, partner_id: partner._id },
            { $set: { match_code: matchCode }, $unset: { user_id: "", partner_id: "" } }
        );
    }
});
```

## Security Considerations

1. **Match Code Validation**: Always validate that a user's match code matches before allowing access
2. **Unmatch Authorization**: Only matched users can initiate unmatch
3. **Data Access**: Users can only access events/photos with their current match code
4. **Historical Data**: Previous relationship data is permanently deleted on unmatch

## Future Enhancements

1. **Soft Delete Option**: Archive shared data instead of hard delete
2. **Match History**: Keep track of past matches (without data)
3. **Data Export**: Allow users to export shared memories before unmatch
4. **Selective Sharing**: Mark some events/photos as personal vs shared
