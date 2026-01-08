package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents the shared post data structure.
type Post struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuthorID  string             `json:"author_id" bson:"author_id" validate:"required"`
	Content   string             `json:"content" bson:"content" validate:"required"`
	MediaURLs []string           `json:"media_urls" bson:"media_urls"`
	Likes     int32              `json:"likes_count" bson:"likes_count"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
