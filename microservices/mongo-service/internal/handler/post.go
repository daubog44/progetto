package handler

import (
	"context"
	"time"

	"github.com/username/progetto/mongo-service/internal/repository"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostHandler struct {
	postv1.UnimplementedPostServiceServer
	repo repository.Repository
}

func NewPostHandler(repo repository.Repository) *PostHandler {
	return &PostHandler{repo: repo}
}

func (h *PostHandler) CreatePost(ctx context.Context, req *postv1.CreatePostRequest) (*postv1.CreatePostResponse, error) {
	doc := repository.PostDoc{
		AuthorID:      req.GetAuthorId(),
		Content:       req.GetContent(),
		MediaURLs:     req.GetMediaUrls(),
		LikesCount:    0,
		CommentsCount: 0,
		CreatedAt:     time.Now(),
	}

	id, err := h.repo.CreatePost(ctx, doc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
	}
	doc.ID = id

	return &postv1.CreatePostResponse{
		Post: h.toProto(doc),
	}, nil
}

func (h *PostHandler) GetPost(ctx context.Context, req *postv1.GetPostRequest) (*postv1.GetPostResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid post id")
	}

	doc, err := h.repo.GetPost(ctx, oid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get post: %v", err)
	}

	return &postv1.GetPostResponse{
		Post: h.toProto(doc),
	}, nil
}

func (h *PostHandler) ListPosts(ctx context.Context, req *postv1.ListPostsRequest) (*postv1.ListPostsResponse, error) {
	limit := int64(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	docs, err := h.repo.ListPosts(ctx, req.GetAuthorId(), limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	var posts []*postv1.Post
	for _, d := range docs {
		posts = append(posts, h.toProto(d))
	}

	return &postv1.ListPostsResponse{
		Posts: posts,
	}, nil
}

func (h *PostHandler) LikePost(ctx context.Context, req *postv1.LikePostRequest) (*postv1.LikePostResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid post id")
	}

	newCount, err := h.repo.LikePost(ctx, oid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to like post: %v", err)
	}

	return &postv1.LikePostResponse{
		Success:       true,
		NewLikesCount: newCount,
	}, nil
}

func (h *PostHandler) toProto(doc repository.PostDoc) *postv1.Post {
	return &postv1.Post{
		Id:            doc.ID.Hex(),
		AuthorId:      doc.AuthorID,
		Content:       doc.Content,
		MediaUrls:     doc.MediaURLs,
		LikesCount:    doc.LikesCount,
		CommentsCount: doc.CommentsCount,
		CreatedAt:     timestamppb.New(doc.CreatedAt),
	}
}
