package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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
	NextPageToken string `query:"next_page" doc:"Token for the next page of results"`
}

type ListPostsOutput struct {
	Body struct {
		Posts         []*postv1.Post `json:"posts"`
		NextPageToken string         `json:"next_page"`
	}
}

func RegisterPostRoutes(api huma.API, client postv1.PostServiceClient, publisher message.Publisher) {
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
			return nil, MapGRPCError(err)
		}

		// Publish Event
		payload, _ := json.Marshal(resp.Post)
		msg := message.NewMessage(watermill.NewUUID(), payload)
		_ = publisher.Publish("post.created", msg)

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
			return nil, MapGRPCError(err)
		}

		// Publish Event
		msg := message.NewMessage(watermill.NewUUID(), []byte(input.ID))
		_ = publisher.Publish("post.liked", msg)

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
