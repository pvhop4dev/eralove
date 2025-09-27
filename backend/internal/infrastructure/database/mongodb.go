package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoDB represents MongoDB connection
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	logger   *zap.Logger
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri, dbName string, logger *zap.Logger) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	logger.Info("Connected to MongoDB", 
		zap.String("database", dbName),
		zap.String("uri", uri))

	return &MongoDB{
		Client:   client,
		Database: database,
		logger:   logger,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.Client.Disconnect(ctx); err != nil {
		m.logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		return err
	}

	m.logger.Info("Disconnected from MongoDB")
	return nil
}

// Collection returns a collection from the database
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// CreateIndexes creates necessary indexes for the collections
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Users collection indexes
	usersCollection := m.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"created_at", 1}},
		},
	}

	if _, err := usersCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Photos collection indexes
	photosCollection := m.Collection("photos")
	photoIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"user_id", 1}, {"date", -1}},
		},
		{
			Keys: bson.D{{"user_id", 1}, {"partner_id", 1}, {"date", -1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	}

	if _, err := photosCollection.Indexes().CreateMany(ctx, photoIndexes); err != nil {
		return fmt.Errorf("failed to create photo indexes: %w", err)
	}

	// Events collection indexes
	eventsCollection := m.Collection("events")
	eventIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"user_id", 1}, {"date", 1}},
		},
		{
			Keys: bson.D{{"user_id", 1}, {"partner_id", 1}, {"date", 1}},
		},
		{
			Keys: bson.D{{"date", 1}},
		},
	}

	if _, err := eventsCollection.Indexes().CreateMany(ctx, eventIndexes); err != nil {
		return fmt.Errorf("failed to create event indexes: %w", err)
	}

	// Match requests collection indexes
	matchRequestsCollection := m.Collection("match_requests")
	matchRequestIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"sender_id", 1}},
		},
		{
			Keys: bson.D{{"receiver_id", 1}},
		},
		{
			Keys: bson.D{{"receiver_email", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	}

	if _, err := matchRequestsCollection.Indexes().CreateMany(ctx, matchRequestIndexes); err != nil {
		return fmt.Errorf("failed to create match request indexes: %w", err)
	}

	m.logger.Info("Database indexes created successfully")
	return nil
}
