package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/username/progetto/shared/pkg/model"
)

type Neo4jRepository struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jRepository(driver neo4j.DriverWithContext) *Neo4jRepository {
	return &Neo4jRepository{driver: driver}
}

// CreatePerson creates a new Person node in Neo4j.
func (r *Neo4jRepository) CreatePerson(ctx context.Context, user model.User) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (p:Person {id: $userID})
			ON CREATE SET p.username = $username, p.email = $email, p.created_at = datetime()
			ON MATCH SET p.username = $username, p.email = $email
			RETURN p
		`
		params := map[string]any{
			"userID":   strconv.Itoa(int(user.ID)),
			"username": user.Username,
			"email":    user.Email,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return result.Collect(ctx)
	})

	if err != nil {
		return fmt.Errorf("failed to create person transaction: %w", err)
	}

	return nil
}
