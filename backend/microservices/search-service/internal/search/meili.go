package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	meilisearchdb "github.com/username/progetto/shared/pkg/database/meilisearch"
)

type SearchResponse struct {
	Hits  []User `json:"hits"`
	Total int64  `json:"total"` // Approximate total
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type MeiliClient struct {
	client meilisearch.ServiceManager
	index  meilisearch.IndexManager
}

func NewMeiliClient(host, key string) *MeiliClient {
	client := meilisearchdb.NewClient(host, key)
	return &MeiliClient{
		client: client,
		index:  client.Index("users"),
	}
}

func (m *MeiliClient) EnsureIndex(ctx context.Context) error {
	// Check if index exists
	_, err := m.client.GetIndexWithContext(ctx, "users")
	if err != nil {
		// Index likely doesn't exist.
		// CreateIndexWithContext
		task, err := m.client.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
			Uid:        "users",
			PrimaryKey: "id",
		})
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		_, err = m.client.WaitForTaskWithContext(ctx, task.TaskUID, 50*time.Millisecond)
		if err != nil {
			return fmt.Errorf("failed to wait for index creation: %w", err)
		}
	}
	return nil
}

func (m *MeiliClient) IndexUser(ctx context.Context, user User) error {
	// AddDocumentsWithContext
	task, err := m.index.AddDocumentsWithContext(ctx, []User{user}, nil)
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	// WaitForTaskWithContext
	_, err = m.client.WaitForTaskWithContext(ctx, task.TaskUID, 50*time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}
	return nil
}

func (m *MeiliClient) SearchUsers(ctx context.Context, query string, limit, offset int64) (*SearchResponse, error) {
	// SearchWithContext
	searchRes, err := m.index.SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	var users []User
	hitsBytes, err := json.Marshal(searchRes.Hits)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hits: %w", err)
	}
	if err := json.Unmarshal(hitsBytes, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hits to users: %w", err)
	}

	return &SearchResponse{
		Hits:  users,
		Total: searchRes.EstimatedTotalHits,
	}, nil
}
