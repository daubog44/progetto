package api

import (
	"context"
	"log/slog"

	searchv1 "github.com/username/progetto/proto/gen/go/search/v1"
	"github.com/username/progetto/search-service/internal/search"
)

type Server struct {
	searchv1.UnimplementedSearchServiceServer
	Meili *search.MeiliClient
}

func NewServer(meili *search.MeiliClient) *Server {
	return &Server{
		Meili: meili,
	}
}

func (s *Server) SearchUsers(ctx context.Context, req *searchv1.SearchUsersRequest) (*searchv1.SearchUsersResponse, error) {
	slog.InfoContext(ctx, "SearchUsers called", "query", req.GetQuery())

	resp, err := s.Meili.SearchUsers(ctx, req.GetQuery(), int64(req.GetLimit()), int64(req.GetOffset()))
	if err != nil {
		slog.ErrorContext(ctx, "failed to search users", "error", err)
		return nil, err
	}

	var users []*searchv1.UserResult
	for _, u := range resp.Hits {
		users = append(users, &searchv1.UserResult{
			Id:       u.ID,
			Username: u.Username,
			Email:    u.Email,
		})
	}

	return &searchv1.SearchUsersResponse{
		Users: users,
		Total: resp.Total,
	}, nil
}
