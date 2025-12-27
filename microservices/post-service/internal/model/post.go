package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID  string             `bson:"author_id"`
	Content   string             `bson:"content"`
	MediaURLs []string           `bson:"media_urls"`
	Likes     int32              `bson:"likes_count"`
	CreatedAt time.Time          `bson:"created_at"`
}

type User struct {
	ID       string `bson:"_id"` // Using Auth Service ID (integer as string) as ID
	Username string `bson:"username"`
	Email    string `bson:"email"`
}
