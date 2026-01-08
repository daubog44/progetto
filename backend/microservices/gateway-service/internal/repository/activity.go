package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserActivity works on the "user_activity" collection in "post-service" db (shared).
type UserActivityRepository struct {
	collection *mongo.Collection
}

func NewUserActivityRepository(db *mongo.Database) *UserActivityRepository {
	return &UserActivityRepository{
		collection: db.Collection("user_activity"),
	}
}

// UpdateLastSeen updates the "last_seen_at" field for a user.
// Uses upsert=true to create the document if it doesn't exist.
func (r *UserActivityRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"last_seen_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
