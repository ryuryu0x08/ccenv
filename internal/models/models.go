// Package models fetches and parses OpenAI-compatible /v1/models responses.
package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Model is one entry from /v1/models.
type Model struct {
	ID            string `json:"id"`
	ContextLength int    `json:"context_length"`
}

type listResponse struct {
	Data []Model `json:"data"`
}

// Fetch GETs an OpenAI-compatible /v1/models URL. If token is non-empty it is
// sent as a Bearer Authorization header. Returns the parsed model list.
func Fetch(url, token string) ([]Model, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models endpoint returned %d: %s", resp.StatusCode, string(body))
	}
	var lr listResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return nil, fmt.Errorf("parse models json: %w", err)
	}
	if len(lr.Data) == 0 {
		return nil, fmt.Errorf("no models returned from %s", url)
	}
	return lr.Data, nil
}
