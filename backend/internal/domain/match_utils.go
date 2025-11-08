package domain

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateMatchCode generates a unique match code for a couple
// The code is deterministic - same two users will always generate the same code
// regardless of who initiates the match
func GenerateMatchCode(userID1, userID2 primitive.ObjectID) string {
	// Convert ObjectIDs to strings
	id1 := userID1.Hex()
	id2 := userID2.Hex()

	// Sort the IDs to ensure consistency regardless of order
	ids := []string{id1, id2}
	sort.Strings(ids)

	// Concatenate sorted IDs
	matchCode := ids[0] + "_" + ids[1]

	return matchCode
}

// ValidateMatchCode checks if a user belongs to a match code
func ValidateMatchCode(matchCode string, userID primitive.ObjectID, partnerID primitive.ObjectID) bool {
	expectedCode := GenerateMatchCode(userID, partnerID)
	return matchCode == expectedCode
}
