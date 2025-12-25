package database

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jClient wraps the Neo4j driver
type Neo4jClient struct {
	driver neo4j.DriverWithContext
}

// NewNeo4jClient creates a new Neo4j client
func NewNeo4jClient(uri, username, password string) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(
		uri,
		neo4j.BasicAuth(username, password, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = 50
			config.MaxConnectionLifetime = 1 * time.Hour
			config.ConnectionAcquisitionTimeout = 2 * time.Minute
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify connectivity: %w", err)
	}

	client := &Neo4jClient{driver: driver}

	// Initialize schema
	if err := client.initializeSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

// Close closes the Neo4j driver
func (c *Neo4jClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.driver.Close(ctx)
}

// GetDriver returns the underlying driver
func (c *Neo4jClient) GetDriver() neo4j.DriverWithContext {
	return c.driver
}

// Session creates a new session
func (c *Neo4jClient) Session(ctx context.Context, mode neo4j.AccessMode) neo4j.SessionWithContext {
	return c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: mode,
	})
}

// ReadSession creates a read-only session
func (c *Neo4jClient) ReadSession(ctx context.Context) neo4j.SessionWithContext {
	return c.Session(ctx, neo4j.AccessModeRead)
}

// WriteSession creates a write session
func (c *Neo4jClient) WriteSession(ctx context.Context) neo4j.SessionWithContext {
	return c.Session(ctx, neo4j.AccessModeWrite)
}

// initializeSchema creates indexes and constraints
func (c *Neo4jClient) initializeSchema(ctx context.Context) error {
	session := c.WriteSession(ctx)
	defer session.Close(ctx)

	constraints := []string{
		// User constraints
		`CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE`,
		`CREATE CONSTRAINT user_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`,

		// Act constraints
		`CREATE CONSTRAINT act_id IF NOT EXISTS FOR (a:Act) REQUIRE a.id IS UNIQUE`,

		// Chain constraints
		`CREATE CONSTRAINT chain_id IF NOT EXISTS FOR (c:Chain) REQUIRE c.id IS UNIQUE`,

		// Testimonial constraints
		`CREATE CONSTRAINT testimonial_id IF NOT EXISTS FOR (t:Testimonial) REQUIRE t.id IS UNIQUE`,
	}

	indexes := []string{
		// User indexes
		`CREATE INDEX user_created_at IF NOT EXISTS FOR (u:User) ON (u.createdAt)`,
		`CREATE INDEX user_location IF NOT EXISTS FOR (u:User) ON (u.location)`,

		// Act indexes
		`CREATE INDEX act_created_at IF NOT EXISTS FOR (a:Act) ON (a.createdAt)`,
		`CREATE INDEX act_type IF NOT EXISTS FOR (a:Act) ON (a.type)`,
		`CREATE INDEX act_status IF NOT EXISTS FOR (a:Act) ON (a.status)`,

		// Chain indexes
		`CREATE INDEX chain_created_at IF NOT EXISTS FOR (c:Chain) ON (c.createdAt)`,

		// Full-text search indexes
		`CREATE FULLTEXT INDEX act_search IF NOT EXISTS FOR (a:Act) ON EACH [a.title, a.description]`,
	}

	// Create constraints
	for _, constraint := range constraints {
		_, err := session.Run(ctx, constraint, nil)
		if err != nil {
			return fmt.Errorf("failed to create constraint: %w", err)
		}
	}

	// Create indexes
	for _, index := range indexes {
		_, err := session.Run(ctx, index, nil)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// ExecuteRead executes a read transaction
func (c *Neo4jClient) ExecuteRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.ReadSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work)
}

// ExecuteWrite executes a write transaction
func (c *Neo4jClient) ExecuteWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.WriteSession(ctx)
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work)
}
