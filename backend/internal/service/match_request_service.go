package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// MatchRequestService implements domain.MatchRequestService
type MatchRequestService struct {
	matchRequestRepo domain.MatchRequestRepository
	userRepo         domain.UserRepository
	logger           *zap.Logger
}

// NewMatchRequestService creates a new match request service
func NewMatchRequestService(
	matchRequestRepo domain.MatchRequestRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.MatchRequestService {
	return &MatchRequestService{
		matchRequestRepo: matchRequestRepo,
		userRepo:         userRepo,
		logger:           logger,
	}
}

// SendMatchRequest sends a match request to another user
func (s *MatchRequestService) SendMatchRequest(
	ctx context.Context,
	senderID primitive.ObjectID,
	req *domain.CreateMatchRequestRequest,
) (*domain.MatchRequestResponse, error) {
	s.logger.Info("Sending match request",
		zap.String("sender_id", senderID.Hex()),
		zap.String("receiver_email", req.ReceiverEmail))

	// Find receiver by email
	receiver, err := s.userRepo.GetByEmail(ctx, req.ReceiverEmail)
	if err != nil {
		s.logger.Error("Receiver not found", zap.Error(err))
		return nil, fmt.Errorf("receiver not found: %w", err)
	}

	// Check if sender is trying to send request to themselves
	if receiver.ID == senderID {
		return nil, fmt.Errorf("cannot send match request to yourself")
	}

	// Check if there's already a pending request
	exists, err := s.matchRequestRepo.ExistsPendingRequest(senderID, receiver.ID)
	if err != nil {
		s.logger.Error("Failed to check pending request", zap.Error(err))
		return nil, fmt.Errorf("failed to check existing requests: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("you already have a pending request to this user")
	}

	// Create match request
	matchRequest := &domain.MatchRequest{
		ID:              primitive.NewObjectID(),
		SenderID:        senderID,
		ReceiverID:      receiver.ID,
		ReceiverEmail:   req.ReceiverEmail,
		AnniversaryDate: req.AnniversaryDate,
		Message:         req.Message,
		Status:          domain.MatchRequestStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.matchRequestRepo.Create(matchRequest); err != nil {
		s.logger.Error("Failed to create match request", zap.Error(err))
		return nil, fmt.Errorf("failed to create match request: %w", err)
	}

	s.logger.Info("Match request sent successfully",
		zap.String("match_request_id", matchRequest.ID.Hex()))

	// Get sender info for response
	sender, _ := s.userRepo.GetByID(ctx, senderID)
	response := matchRequest.ToResponse()
	if sender != nil {
		response.SenderName = sender.Name
		response.SenderEmail = sender.Email
	}

	return response, nil
}

// GetMatchRequest gets a specific match request
func (s *MatchRequestService) GetMatchRequest(
	ctx context.Context,
	requestID, userID primitive.ObjectID,
) (*domain.MatchRequestResponse, error) {
	matchRequest, err := s.matchRequestRepo.GetByID(requestID)
	if err != nil {
		return nil, err
	}

	// Verify user has access
	if matchRequest.SenderID != userID && matchRequest.ReceiverID != userID {
		return nil, fmt.Errorf("unauthorized access to match request")
	}

	response := matchRequest.ToResponse()

	// Get sender info
	sender, _ := s.userRepo.GetByID(ctx, matchRequest.SenderID)
	if sender != nil {
		response.SenderName = sender.Name
		response.SenderEmail = sender.Email
	}

	return response, nil
}

// GetSentRequests gets match requests sent by a user
func (s *MatchRequestService) GetSentRequests(
	ctx context.Context,
	userID primitive.ObjectID,
	status string,
	page, limit int,
) ([]*domain.MatchRequestResponse, int64, error) {
	s.logger.Info("Getting sent match requests",
		zap.String("user_id", userID.Hex()))

	offset := (page - 1) * limit
	matchRequests, err := s.matchRequestRepo.GetBySenderID(userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get sent requests", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get sent requests: %w", err)
	}

	// Filter by status if provided
	var filtered []*domain.MatchRequest
	if status != "" {
		for _, mr := range matchRequests {
			if string(mr.Status) == status {
				filtered = append(filtered, mr)
			}
		}
	} else {
		filtered = matchRequests
	}

	responses := make([]*domain.MatchRequestResponse, len(filtered))
	for i, mr := range filtered {
		responses[i] = mr.ToResponse()
	}

	return responses, int64(len(filtered)), nil
}

// GetReceivedRequests gets match requests received by a user
func (s *MatchRequestService) GetReceivedRequests(
	ctx context.Context,
	userID primitive.ObjectID,
	status string,
	page, limit int,
) ([]*domain.MatchRequestResponse, int64, error) {
	s.logger.Info("Getting received match requests",
		zap.String("user_id", userID.Hex()),
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("limit", limit))

	offset := (page - 1) * limit
	matchRequests, err := s.matchRequestRepo.GetByReceiverID(userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get received requests", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get received requests: %w", err)
	}

	s.logger.Info("Retrieved match requests from DB",
		zap.Int("total_count", len(matchRequests)))

	// Filter by status if provided
	var filtered []*domain.MatchRequest
	if status != "" {
		for _, mr := range matchRequests {
			if string(mr.Status) == status {
				filtered = append(filtered, mr)
			}
		}
		s.logger.Info("Filtered by status",
			zap.String("status", status),
			zap.Int("filtered_count", len(filtered)))
	} else {
		filtered = matchRequests
	}

	responses := make([]*domain.MatchRequestResponse, len(filtered))
	for i, mr := range filtered {
		response := mr.ToResponse()
		
		// Get sender info
		sender, err := s.userRepo.GetByID(ctx, mr.SenderID)
		if err == nil && sender != nil {
			response.SenderName = sender.Name
			response.SenderEmail = sender.Email
			s.logger.Info("Populated sender info",
				zap.String("request_id", mr.ID.Hex()),
				zap.String("sender_name", sender.Name),
				zap.String("sender_email", sender.Email))
		} else {
			s.logger.Warn("Failed to get sender info",
				zap.String("request_id", mr.ID.Hex()),
				zap.String("sender_id", mr.SenderID.Hex()),
				zap.Error(err))
		}
		
		responses[i] = response
	}

	s.logger.Info("Returning received requests",
		zap.Int("response_count", len(responses)))

	return responses, int64(len(filtered)), nil
}

// RespondToMatchRequest responds to a match request (accept/reject)
func (s *MatchRequestService) RespondToMatchRequest(
	ctx context.Context,
	requestID, userID primitive.ObjectID,
	req *domain.RespondToMatchRequestRequest,
) (*domain.MatchRequestResponse, error) {
	s.logger.Info("Responding to match request",
		zap.String("request_id", requestID.Hex()),
		zap.String("user_id", userID.Hex()),
		zap.String("action", req.Action))

	// Get the match request
	matchRequest, err := s.matchRequestRepo.GetByID(requestID)
	if err != nil {
		s.logger.Error("Match request not found", zap.Error(err))
		return nil, fmt.Errorf("match request not found: %w", err)
	}

	// Verify that the user is the receiver
	if matchRequest.ReceiverID != userID {
		return nil, fmt.Errorf("unauthorized to respond to this request")
	}

	// Check if already responded
	if matchRequest.Status != domain.MatchRequestStatusPending {
		return nil, fmt.Errorf("match request already responded to")
	}

	// Update status based on action
	now := time.Now()
	matchRequest.RespondedAt = &now
	matchRequest.UpdatedAt = now

	if req.Action == "accept" {
		matchRequest.Status = domain.MatchRequestStatusAccepted
		
		// Generate match code for the couple
		matchCode := domain.GenerateMatchCode(matchRequest.SenderID, matchRequest.ReceiverID)
		
		// Get both users
		sender, err := s.userRepo.GetByID(ctx, matchRequest.SenderID)
		if err != nil {
			s.logger.Error("Failed to get sender", zap.Error(err))
			return nil, fmt.Errorf("failed to get sender: %w", err)
		}
		
		receiver, err := s.userRepo.GetByID(ctx, matchRequest.ReceiverID)
		if err != nil {
			s.logger.Error("Failed to get receiver", zap.Error(err))
			return nil, fmt.Errorf("failed to get receiver: %w", err)
		}
		
		// Determine which anniversary date to use
		// Priority: 1. Receiver's override, 2. Original request date
		var finalAnniversaryDate time.Time
		if req.AnniversaryDate != nil {
			// Receiver provided a different date when accepting
			finalAnniversaryDate = *req.AnniversaryDate
			s.logger.Info("Using receiver's anniversary date",
				zap.Time("anniversary_date", finalAnniversaryDate))
		} else {
			// Use the date from original match request
			finalAnniversaryDate = matchRequest.AnniversaryDate
			s.logger.Info("Using sender's anniversary date",
				zap.Time("anniversary_date", finalAnniversaryDate))
		}
		
		// Update sender with match info
		sender.PartnerID = &matchRequest.ReceiverID
		sender.PartnerName = receiver.Name
		sender.MatchCode = matchCode
		sender.MatchedAt = &now
		sender.AnniversaryDate = &finalAnniversaryDate
		sender.UpdatedAt = now
		
		if err := s.userRepo.Update(ctx, sender.ID, sender); err != nil {
			s.logger.Error("Failed to update sender", zap.Error(err))
			return nil, fmt.Errorf("failed to update sender: %w", err)
		}
		
		// Update receiver with match info
		receiver.PartnerID = &matchRequest.SenderID
		receiver.PartnerName = sender.Name
		receiver.MatchCode = matchCode
		receiver.MatchedAt = &now
		receiver.AnniversaryDate = &finalAnniversaryDate
		receiver.UpdatedAt = now
		
		if err := s.userRepo.Update(ctx, receiver.ID, receiver); err != nil {
			s.logger.Error("Failed to update receiver", zap.Error(err))
			return nil, fmt.Errorf("failed to update receiver: %w", err)
		}
		
		s.logger.Info("Match created successfully",
			zap.String("match_code", matchCode),
			zap.String("sender_id", sender.ID.Hex()),
			zap.String("receiver_id", receiver.ID.Hex()),
			zap.Time("anniversary_date", finalAnniversaryDate))
	} else {
		matchRequest.Status = domain.MatchRequestStatusDeclined
	}

	if err := s.matchRequestRepo.Update(requestID, matchRequest); err != nil {
		s.logger.Error("Failed to update match request", zap.Error(err))
		return nil, fmt.Errorf("failed to update match request: %w", err)
	}

	s.logger.Info("Match request responded successfully",
		zap.String("request_id", requestID.Hex()),
		zap.String("status", string(matchRequest.Status)))

	// Get sender info for response
	sender, _ := s.userRepo.GetByID(ctx, matchRequest.SenderID)
	response := matchRequest.ToResponse()
	if sender != nil {
		response.SenderName = sender.Name
		response.SenderEmail = sender.Email
	}

	return response, nil
}

// CancelMatchRequest cancels a sent match request
func (s *MatchRequestService) CancelMatchRequest(
	ctx context.Context,
	requestID, userID primitive.ObjectID,
) error {
	s.logger.Info("Canceling match request",
		zap.String("request_id", requestID.Hex()),
		zap.String("user_id", userID.Hex()))

	// Get the match request
	matchRequest, err := s.matchRequestRepo.GetByID(requestID)
	if err != nil {
		s.logger.Error("Match request not found", zap.Error(err))
		return fmt.Errorf("match request not found: %w", err)
	}

	// Verify that the user is the sender
	if matchRequest.SenderID != userID {
		return fmt.Errorf("unauthorized to cancel this request")
	}

	// Can only cancel pending requests
	if matchRequest.Status != domain.MatchRequestStatusPending {
		return fmt.Errorf("can only cancel pending requests")
	}

	if err := s.matchRequestRepo.Delete(requestID); err != nil {
		s.logger.Error("Failed to delete match request", zap.Error(err))
		return fmt.Errorf("failed to cancel match request: %w", err)
	}

	s.logger.Info("Match request canceled successfully",
		zap.String("request_id", requestID.Hex()))

	return nil
}
