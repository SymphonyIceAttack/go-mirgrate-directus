// Package gomirgratedirectus provides functions to migrate Directus schema between environments.
package gomirgratedirectus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DirectusClient holds the configuration for a Directus instance.
type DirectusClient struct {
	URL         string
	AccessToken string
	HTTPClient  *http.Client
}

// NewDirectusClient creates a new client for a Directus instance.
func NewDirectusClient(url, accessToken string) *DirectusClient {
	return &DirectusClient{
		URL:         url,
		AccessToken: accessToken,
		HTTPClient:  &http.Client{},
	}
}

// GetSnapshot retrieves a schema snapshot from the Directus instance.
func (c *DirectusClient) GetSnapshot() (map[string]any, error) {
	url := fmt.Sprintf("%s/schema/snapshot?access_token=%s", c.URL, c.AccessToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute snapshot request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("snapshot request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode snapshot response: %w", err)
	}

	if data, ok := result["data"]; ok {
		return data.(map[string]any), nil
	}

	return nil, fmt.Errorf("snapshot response does not contain 'data' field")
}

// GetDiff retrieves a schema diff between the target instance and the provided snapshot.
func (c *DirectusClient) GetDiff(snapshot map[string]any, force bool) (map[string]any, error) {
	url := fmt.Sprintf("%s/schema/diff?access_token=%s", c.URL, c.AccessToken)
	if force {
		url += "&force=true"
	}

	requestBody, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot for diff request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create diff request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute diff request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("diff request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode diff response: %w", err)
	}

	if data, ok := result["data"]; ok {
		return data.(map[string]any), nil
	}

	return nil, fmt.Errorf("diff response does not contain 'data' field")
}

// ApplyDiff applies a schema diff to the Directus instance.
func (c *DirectusClient) ApplyDiff(diff map[string]any) error {
	url := fmt.Sprintf("%s/schema/apply?access_token=%s", c.URL, c.AccessToken)

	requestBody, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("failed to marshal diff for apply request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create apply request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute apply request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apply request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Migrate performs a full schema migration from a base project to a target project.
func Migrate(baseURL, baseToken, targetURL, targetToken string, force bool) error {
	baseClient := NewDirectusClient(baseURL, baseToken)
	targetClient := NewDirectusClient(targetURL, targetToken)

	fmt.Println("Retrieving snapshot from base project...")
	snapshot, err := baseClient.GetSnapshot()
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}
	fmt.Println("Snapshot retrieved successfully.")

	fmt.Println("Retrieving diff from target project...")
	diff, err := targetClient.GetDiff(snapshot, force)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}
	fmt.Println("Diff retrieved successfully.")

	fmt.Println("Applying diff to target project...")
	if err := targetClient.ApplyDiff(diff); err != nil {
		return fmt.Errorf("failed to apply diff: %w", err)
	}
	fmt.Println("Diff applied successfully. Migration complete.")

	return nil
}
