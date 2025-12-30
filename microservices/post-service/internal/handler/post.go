package handler

import (
	"context"
	"encoding/json"
	"time"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/post-service/internal/model"
	"github.com/username/progetto/post-service/internal/repository"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostHandler struct {
	postv1.UnimplementedPostServiceServer
	repo      repository.PostRepository
	publisher message.Publisher
	logger    *slog.Logger
}

func NewPostHandler(repo repository.PostRepository, publisher message.Publisher) *PostHandler {
	return &PostHandler{
		repo:      repo,
		publisher: publisher,
		logger:    slog.Default().With("component", "post_handler"),
	}
}

func (h *PostHandler) CreatePost(ctx context.Context, req *postv1.CreatePostRequest) (*postv1.CreatePostResponse, error) {
	if req.AuthorId == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "author_id and content are required")
	}

	post := &model.Post{
		AuthorID:  req.AuthorId,
		Content:   req.Content,
		MediaURLs: req.MediaUrls,
		Likes:     0,
		CreatedAt: time.Now(),
	}

	if err := h.repo.Create(ctx, post); err != nil {
		h.logger.ErrorContext(ctx, "failed to create post", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
	}

	// Publish Event
	protoPost := h.mapToProto(post)
	payload, _ := json.Marshal(protoPost)
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)
	if err := h.publisher.Publish("post.created", msg); err != nil {
		// Log error but proceed
		h.logger.ErrorContext(ctx, "failed to publish post.created event", "error", err, "post_id", post.ID.Hex())
	}

	return &postv1.CreatePostResponse{
		Post: protoPost,
	}, nil
}

func (h *PostHandler) GetPost(ctx context.Context, req *postv1.GetPostRequest) (*postv1.GetPostResponse, error) {
	post, err := h.repo.GetByID(ctx, req.PostId)
	if err != nil {
		h.logger.WarnContext(ctx, "post not found", "error", err, "post_id", req.PostId)
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}
	return &postv1.GetPostResponse{
		Post: h.mapToProto(post),
	}, nil
}

func (h *PostHandler) ListPosts(ctx context.Context, req *postv1.ListPostsRequest) (*postv1.ListPostsResponse, error) {
	limit := int64(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	posts, nextToken, err := h.repo.List(ctx, req.AuthorId, limit, req.NextPageToken)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list posts", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	var protoPosts []*postv1.Post
	for _, p := range posts {
		protoPosts = append(protoPosts, h.mapToProto(p))
	}

	return &postv1.ListPostsResponse{
		Posts:         protoPosts,
		NextPageToken: nextToken,
	}, nil
}

func (h *PostHandler) mapToProto(p *model.Post) *postv1.Post {
	return &postv1.Post{
		Id:         p.ID.Hex(),
		AuthorId:   p.AuthorID,
		Content:    p.Content,
		MediaUrls:  p.MediaURLs,
		LikesCount: p.Likes,
		CreatedAt:  timestamppb.New(p.CreatedAt),
	}
}
