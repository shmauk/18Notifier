package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// DGraphGraphQLAdapter implements GraphQLAdapter for DGraph
type DGraphGraphQLAdapter struct {
	endpoint string
	client   *http.Client
	timeout  time.Duration
}

// NewDGraphGraphQLAdapter creates a new DGraph GraphQL adapter
func NewDGraphGraphQLAdapter(endpoint string) *DGraphGraphQLAdapter {
	return &DGraphGraphQLAdapter{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 30 * time.Second},
		timeout:  30 * time.Second,
	}
}

// Query executes a GraphQL query against DGraph
func (a *DGraphGraphQLAdapter) Query(query string, variables map[string]interface{}) ([]byte, error) {
	return a.executeGraphQL(query, variables)
}

// Mutate executes a GraphQL mutation against DGraph
func (a *DGraphGraphQLAdapter) Mutate(mutation string, variables map[string]interface{}) ([]byte, error) {
	return a.executeGraphQL(mutation, variables)
}

// executeGraphQL executes a GraphQL operation against DGraph with retry logic
func (a *DGraphGraphQLAdapter) executeGraphQL(operation string, variables map[string]interface{}) ([]byte, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying GraphQL operation (attempt %d/%d)", attempt+1, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		result, err := a.executeSingleGraphQL(operation, variables)
		if err != nil {
			lastErr = err
			log.Printf("GraphQL operation failed (attempt %d): %v", attempt+1, err)
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("GraphQL operation failed after %d attempts: %w", maxRetries, lastErr)
}

// executeSingleGraphQL executes a single GraphQL operation
func (a *DGraphGraphQLAdapter) executeSingleGraphQL(operation string, variables map[string]interface{}) ([]byte, error) {
	requestBody := map[string]interface{}{
		"query":     operation,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", a.endpoint+"/graphql", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for GraphQL errors
	if errors, ok := result["errors"]; ok && errors != nil {
		return nil, fmt.Errorf("GraphQL errors: %v", errors)
	}

	responseJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return responseJSON, nil
}

// HealthCheck checks if the DGraph GraphQL endpoint is healthy
func (a *DGraphGraphQLAdapter) HealthCheck() error {
	query := `
		query HealthCheck {
			__schema {
				types {
					name
				}
			}
		}
	`

	_, err := a.Query(query, nil)
	if err != nil {
		return fmt.Errorf("DGraph health check failed: %w", err)
	}

	return nil
}

// GetEndpoint returns the GraphQL endpoint URL
func (a *DGraphGraphQLAdapter) GetEndpoint() string {
	return a.endpoint
}

// SetTimeout sets the timeout for GraphQL operations
func (a *DGraphGraphQLAdapter) SetTimeout(timeout time.Duration) {
	a.timeout = timeout
	a.client.Timeout = timeout
}
