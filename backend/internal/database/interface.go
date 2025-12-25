package database

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// DBClient defines the interface for database operations
type DBClient interface {
	ExecuteRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error)
	ExecuteWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error)
	ReadSession(ctx context.Context) neo4j.SessionWithContext
	Close() error
}

// Ensure Neo4jClient implements DBClient
var _ DBClient = (*Neo4jClient)(nil)
