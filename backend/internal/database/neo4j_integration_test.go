package database

import (
	"context"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupNeo4jContainer(t *testing.T) (*Neo4jClient, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "neo4j:5.15",
		ExposedPorts: []string{"7687/tcp"},
		Env: map[string]string{
			"NEO4J_AUTH": "neo4j/testpassword",
		},
		WaitingFor: wait.ForLog("Started").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "7687")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	uri := "bolt://" + host + ":" + mappedPort.Port()

	time.Sleep(2 * time.Second)

	client, err := NewNeo4jClient(uri, "neo4j", "testpassword")
	if err != nil {
		t.Fatalf("Failed to create Neo4j client: %v", err)
	}

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestNeo4jClient_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, cleanup := setupNeo4jContainer(t)
	defer cleanup()

	ctx := context.Background()

	result, err := client.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return "connected", nil
	})

	if err != nil {
		t.Errorf("Failed to execute read: %v", err)
	}

	if result != "connected" {
		t.Errorf("Expected 'connected', got %v", result)
	}
}

func TestNeo4jClient_CreateAndReadNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, cleanup := setupNeo4jContainer(t)
	defer cleanup()

	ctx := context.Background()

	testID := "test-id-123"
	testName := "Test User"

	_, err := client.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `CREATE (u:User {id: $id, name: $name}) RETURN u`
		params := map[string]interface{}{
			"id":   testID,
			"name": testName,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	result, err := client.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (u:User {id: $id}) RETURN u.name as name`
		params := map[string]interface{}{
			"id": testID,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			name, _ := record.Get("name")
			return name, nil
		}

		return nil, nil
	})

	if err != nil {
		t.Fatalf("Failed to read node: %v", err)
	}

	if result != testName {
		t.Errorf("Expected name %s, got %v", testName, result)
	}
}

func TestNeo4jClient_SchemaInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, cleanup := setupNeo4jContainer(t)
	defer cleanup()

	ctx := context.Background()

	result, err := client.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `SHOW CONSTRAINTS`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var constraints []string
		for result.Next(ctx) {
			record := result.Record()
			if name, ok := record.Get("name"); ok {
				constraints = append(constraints, name.(string))
			}
		}

		return constraints, nil
	})

	if err != nil {
		t.Fatalf("Failed to get constraints: %v", err)
	}

	constraints := result.([]string)
	if len(constraints) == 0 {
		t.Error("Expected constraints to be created, but found none")
	}
}

func TestNeo4jClient_Transaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, cleanup := setupNeo4jContainer(t)
	defer cleanup()

	ctx := context.Background()

	testID1 := "test-tx-1"
	testID2 := "test-tx-2"

	_, err := client.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query1 := `CREATE (u:TestUser {id: $id, name: 'User 1'})`
		_, err := tx.Run(ctx, query1, map[string]interface{}{"id": testID1})
		if err != nil {
			return nil, err
		}

		query2 := `CREATE (u:TestUser {id: $id, name: 'User 2'})`
		_, err = tx.Run(ctx, query2, map[string]interface{}{"id": testID2})
		if err != nil {
			return nil, err
		}

		return "success", nil
	})

	if err != nil {
		t.Fatalf("Failed to execute transaction: %v", err)
	}

	count, err := client.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (u:TestUser) RETURN count(u) as count`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			return count, nil
		}

		return int64(0), nil
	})

	if err != nil {
		t.Fatalf("Failed to count nodes: %v", err)
	}

	if count != int64(2) {
		t.Errorf("Expected 2 nodes, got %v", count)
	}
}
