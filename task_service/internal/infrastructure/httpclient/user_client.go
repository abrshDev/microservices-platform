package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UserClient talks to user-service over REST, hitting its internal-only
// endpoints. This is a parallel path to the existing gRPC UserClient —
// it doesn't replace it, just demonstrates the REST-internal pattern.
type UserClient struct {
	baseURL string // e.g. "http://user-service:8080"
	apiKey  string // shared INTERNAL_API_KEY, same value user-service checks
}

func NewUserClient(baseURL, apiKey string) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// addAuth injects the internal API key into every outgoing request.
// Written once, called by every method below.
func (c *UserClient) addAuth(req *http.Request) {
	req.Header.Set("X-INTERNAL-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

// GetUserByID fetches a single user by ID — path param, since ID
// identifies exactly which resource we want.
func (c *UserClient) GetUserByID(ctx context.Context, userID string) (*User, error) {
	url := fmt.Sprintf("%s/api/v1/internal/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	c.addAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user %s not found", userID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}

// ListUsersByStatus fetches users filtered by status — query param,
// since status is an optional filter, not an identifier.
func (c *UserClient) ListUsersByStatus(ctx context.Context, status string) ([]User, error) {
	url := fmt.Sprintf("%s/api/v1/internal/users?status=%s", c.baseURL, status)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	c.addAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data []User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// User is the response shape from user-service's internal endpoints.
// Only the fields task-service actually needs — not the full domain model.
type User struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}
