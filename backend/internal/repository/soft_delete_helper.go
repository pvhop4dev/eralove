package repository

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SoftDeleteFilter provides helper functions for soft delete operations
type SoftDeleteFilter struct{}

// NewSoftDeleteFilter creates a new soft delete filter helper
func NewSoftDeleteFilter() *SoftDeleteFilter {
	return &SoftDeleteFilter{}
}

// GetActiveFilter returns a filter for active (non-deleted) documents
func (s *SoftDeleteFilter) GetActiveFilter() bson.M {
	return bson.M{
		"deleted_at": bson.M{"$exists": false},
	}
}

// GetActiveFilterWithCondition returns a filter for active documents with additional conditions
func (s *SoftDeleteFilter) GetActiveFilterWithCondition(condition bson.M) bson.M {
	filter := s.GetActiveFilter()
	for k, v := range condition {
		filter[k] = v
	}
	return filter
}

// GetDeletedFilter returns a filter for deleted documents
func (s *SoftDeleteFilter) GetDeletedFilter() bson.M {
	return bson.M{
		"deleted_at": bson.M{"$exists": true},
	}
}

// GetDeletedFilterWithCondition returns a filter for deleted documents with additional conditions
func (s *SoftDeleteFilter) GetDeletedFilterWithCondition(condition bson.M) bson.M {
	filter := s.GetDeletedFilter()
	for k, v := range condition {
		filter[k] = v
	}
	return filter
}

// CreateSoftDeleteUpdate creates an update document for soft delete
func (s *SoftDeleteFilter) CreateSoftDeleteUpdate() bson.M {
	now := time.Now()
	return bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}
}

// CreateRestoreUpdate creates an update document for restoring soft deleted document
func (s *SoftDeleteFilter) CreateRestoreUpdate() bson.M {
	return bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"deleted_at": "",
		},
	}
}

// CreateSoftDeleteUpdateWithFields creates an update document for soft delete with additional fields
func (s *SoftDeleteFilter) CreateSoftDeleteUpdateWithFields(additionalFields bson.M) bson.M {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}
	
	// Add additional fields to the $set operation
	for k, v := range additionalFields {
		update["$set"].(bson.M)[k] = v
	}
	
	return update
}

// GetActiveFilterByUserID returns a filter for active documents by user ID
func (s *SoftDeleteFilter) GetActiveFilterByUserID(userID primitive.ObjectID) bson.M {
	return s.GetActiveFilterWithCondition(bson.M{"user_id": userID})
}

// GetActiveFilterByID returns a filter for active document by ID
func (s *SoftDeleteFilter) GetActiveFilterByID(id primitive.ObjectID) bson.M {
	return s.GetActiveFilterWithCondition(bson.M{"_id": id})
}

// GetActiveFilterByCoupleID returns a filter for active documents by couple (user and partner)
func (s *SoftDeleteFilter) GetActiveFilterByCoupleID(userID, partnerID primitive.ObjectID) bson.M {
	return s.GetActiveFilterWithCondition(bson.M{
		"$or": []bson.M{
			{"user_id": userID, "partner_id": partnerID},
			{"user_id": partnerID, "partner_id": userID},
		},
	})
}

// GetActiveFilterByConversation returns a filter for active messages in a conversation
func (s *SoftDeleteFilter) GetActiveFilterByConversation(userID, partnerID primitive.ObjectID) bson.M {
	return s.GetActiveFilterWithCondition(bson.M{
		"$or": []bson.M{
			{"sender_id": userID, "receiver_id": partnerID},
			{"sender_id": partnerID, "receiver_id": userID},
		},
	})
}

// Global instance for easy access
var SoftDelete = NewSoftDeleteFilter()
