package repository

import (
	"context"

	"github.com/username/progetto/post-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, authorID string, limit int64, cursor string) ([]*model.Post, string, error)
}

type UserRepository interface {
	Save(ctx context.Context, user *model.User) error
}

type mongoPostRepository struct {
	collection *mongo.Collection
}

func NewMongoPostRepository(db *mongo.Database) PostRepository {
	return &mongoPostRepository{collection: db.Collection("posts")}
}

func (r *mongoPostRepository) Create(ctx context.Context, post *model.Post) error {
	res, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return err
	}
	post.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoPostRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var post model.Post
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&post)
	return &post, err
}

func (r *mongoPostRepository) List(ctx context.Context, authorID string, limit int64, cursor string) ([]*model.Post, string, error) {
	filter := bson.M{}
	if authorID != "" {
		filter["author_id"] = authorID
	}

	// Pagination: if cursor is present, fetch items older than the cursor (represented by ObjectID)
	if cursor != "" {
		oid, err := primitive.ObjectIDFromHex(cursor)
		if err == nil {
			filter["_id"] = bson.M{"$lt": oid}
		}
	}

	opts := options.Find().SetLimit(limit).SetSort(bson.M{"_id": -1}) // Sort by ID desc (newest first)
	cur, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	var posts []*model.Post
	if err := cur.All(ctx, &posts); err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursor = lastPost.ID.Hex()
	}

	return posts, nextCursor, nil
}

type mongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{collection: db.Collection("users")}
}

func (r *mongoUserRepository) Save(ctx context.Context, user *model.User) error {
	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user, opts)
	return err
}
