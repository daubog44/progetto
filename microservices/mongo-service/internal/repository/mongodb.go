package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID      string             `bson:"author_id"`
	Content       string             `bson:"content"`
	MediaURLs     []string           `bson:"media_urls"`
	LikesCount    int32              `bson:"likes_count"`
	CommentsCount int32              `bson:"comments_count"`
	CreatedAt     time.Time          `bson:"created_at"`
}

type Repository interface {
	CreatePost(ctx context.Context, doc PostDoc) (primitive.ObjectID, error)
	GetPost(ctx context.Context, id primitive.ObjectID) (PostDoc, error)
	ListPosts(ctx context.Context, authorID string, limit int64) ([]PostDoc, error)
	LikePost(ctx context.Context, id primitive.ObjectID) (int32, error)
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(collection *mongo.Collection) Repository {
	return &mongoRepository{collection: collection}
}

func (r *mongoRepository) CreatePost(ctx context.Context, doc PostDoc) (primitive.ObjectID, error) {
	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *mongoRepository) GetPost(ctx context.Context, id primitive.ObjectID) (PostDoc, error) {
	var doc PostDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	return doc, err
}

func (r *mongoRepository) ListPosts(ctx context.Context, authorID string, limit int64) ([]PostDoc, error) {
	filter := bson.M{}
	if authorID != "" {
		filter["author_id"] = authorID
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []PostDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *mongoRepository) LikePost(ctx context.Context, id primitive.ObjectID) (int32, error) {
	update := bson.M{"$inc": bson.M{"likes_count": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc PostDoc
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&doc)
	return doc.LikesCount, err
}
