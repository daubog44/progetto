package repository

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepository struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jRepository(driver neo4j.DriverWithContext) *Neo4jRepository {
	return &Neo4jRepository{driver: driver}
}

// CreatePerson creates a new Person node in Neo4j.
func (r *Neo4jRepository) CreatePerson(ctx context.Context, userID, username, email string) error {
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
			"userID":   userID,
			"username": username,
			"email":    email,
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
