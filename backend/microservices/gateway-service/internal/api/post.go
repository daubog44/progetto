package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostInput struct {
	Body struct {
		AuthorID  string   `json:"author_id" doc:"Author of the post"`
		Content   string   `json:"content" doc:"Content of the post"`
		MediaURLs []string `json:"media_urls" doc:"List of media URLs"`
	}
}

type PostOutput struct {
	Body struct {
		Post *postv1.Post `json:"post"`
	}
}

type ListPostsInput struct {
	AuthorID      string `query:"author_id" doc:"Filter by author ID"`
	Limit         int32  `query:"limit" doc:"Maximum number of posts to return" default:"20"`
	NextPageToken string `query:"anchorPage" doc:"Token for the next page of results"`
}

type ListPostsOutput struct {
	Body struct {
		Posts         []*postv1.Post `json:"posts"`
		NextPageToken string         `json:"anchorPage"`
	}
}

func RegisterPostRoutes(api huma.API, client postv1.PostServiceClient, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "create-post",
		Method:      http.MethodPost,
		Path:        "/posts",
		Summary:     "Create a post",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, input *PostInput) (*PostOutput, error) {
		resp, err := client.CreatePost(ctx, &postv1.CreatePostRequest{
			AuthorId:  input.Body.AuthorID,
			Content:   input.Body.Content,
			MediaUrls: input.Body.MediaURLs,
		})
		if err != nil {
			logger.ErrorContext(ctx, "create post failed", "error", err)
			return nil, MapGRPCError(err)
		}

		output := &PostOutput{}
		output.Body.Post = resp.Post
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-post",
		Method:      http.MethodGet,
		Path:        "/posts/{id}",
		Summary:     "Get a post",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id"`
	}) (*PostOutput, error) {
		resp, err := client.GetPost(ctx, &postv1.GetPostRequest{
			PostId: input.ID,
		})
		if err != nil {
			logger.WarnContext(ctx, "get post failed", "error", err, "post_id", input.ID) // Warn because likely 404
			return nil, MapGRPCError(err)
		}

		output := &PostOutput{}
		output.Body.Post = resp.Post
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-posts",
		Method:      http.MethodGet,
		Path:        "/posts",
		Summary:     "List posts",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, input *ListPostsInput) (*ListPostsOutput, error) {
		resp, err := client.ListPosts(ctx, &postv1.ListPostsRequest{
			AuthorId:      input.AuthorID,
			Limit:         input.Limit,
			NextPageToken: input.NextPageToken,
		})
		if err != nil {
			logger.ErrorContext(ctx, "list posts failed", "error", err)
			return nil, MapGRPCError(err)
		}

		output := &ListPostsOutput{}
		output.Body.Posts = resp.Posts
		output.Body.NextPageToken = resp.NextPageToken
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "like-post",
		Method:      http.MethodPost,
		Path:        "/posts/{id}/like",
		Summary:     "Like a post",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, input *struct {
		ID     string `path:"id"`
		UserID string `json:"user_id"`
	}) (*struct {
		Body struct {
			Success bool `json:"success"`
		}
	}, error) {
		_, err := client.LikePost(ctx, &postv1.LikePostRequest{
			PostId: input.ID,
			UserId: input.UserID,
		})
		if err != nil {
			logger.ErrorContext(ctx, "like post failed", "error", err)
			return nil, MapGRPCError(err)
		}

		return &struct {
			Body struct {
				Success bool `json:"success"`
			}
		}{Body: struct {
			Success bool `json:"success"`
		}{Success: true}}, nil
	})
}

func MapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return huma.Error500InternalServerError("Unexpected error", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return huma.Error404NotFound(st.Message())
	case codes.InvalidArgument:
		return huma.Error400BadRequest(st.Message())
	case codes.Unauthenticated:
		return huma.Error401Unauthorized(st.Message())
	case codes.PermissionDenied:
		return huma.Error403Forbidden(st.Message())
	default:
		return huma.Error500InternalServerError(st.Message())
	}
}
