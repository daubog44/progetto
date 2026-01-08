package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	searchv1 "github.com/username/progetto/proto/gen/go/search/v1"
)

type SearchUsersInput struct {
	Q      string `query:"q" doc:"Search query" required:"true"`
	Limit  int32  `query:"limit" doc:"Limit results" default:"10"`
	Offset int32  `query:"offset" doc:"Offset results" default:"0"`
}

type SearchUsersOutput struct {
	Body struct {
		Users []UserResult `json:"users"`
		Total int64        `json:"total"`
	}
}

type UserResult struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func RegisterSearchRoutes(api huma.API, searchClient searchv1.SearchServiceClient, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "search-users",
		Method:      http.MethodGet,
		Path:        "/search-users",
		Summary:     "Search users",
		Tags:        []string{"Search"},
	}, func(ctx context.Context, input *SearchUsersInput) (*SearchUsersOutput, error) {
		resp, err := searchClient.SearchUsers(ctx, &searchv1.SearchUsersRequest{
			Query:  input.Q,
			Limit:  input.Limit,
			Offset: input.Offset,
		})
		if err != nil {
			logger.Error("failed to search users", "error", err)
			return nil, huma.Error500InternalServerError("failed to search users")
		}

		users := make([]UserResult, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, UserResult{
				ID:       u.Id,
				Username: u.Username,
				Email:    u.Email,
			})
		}

		return &SearchUsersOutput{
			Body: struct {
				Users []UserResult `json:"users"`
				Total int64        `json:"total"`
			}{
				Users: users,
				Total: resp.Total,
			},
		}, nil
	})
}
