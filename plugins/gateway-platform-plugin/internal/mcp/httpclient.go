package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Key struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type Route struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	LocalPath   string `json:"local_path"`
	UpstreamURL string `json:"upstream_url"`
	TimeoutMS   int    `json:"timeout_ms"`
	Description string `json:"description"`
}

type ReferenceReport struct {
	MissingKeys []string `json:"missing_keys"`
	UnusedKeys  []string `json:"unused_keys"`
}

type HealthStatus struct {
	OK               bool   `json:"ok"`
	DataDir          string `json:"data_dir"`
	DatabasePath     string `json:"database_path"`
	DatabaseWritable bool   `json:"database_writable"`
	PID              int    `json:"pid"`
	ExecutablePath   string `json:"executable_path"`
	Version          string `json:"version"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}
}

func (c *Client) Health() (*HealthStatus, error) {
	resp, err := c.http.Get(c.baseURL + "/api/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return &status, fmt.Errorf("health check failed: %s", resp.Status)
	}
	return &status, nil
}

func (c *Client) HealthCheck() (bool, error) {
	status, err := c.Health()
	if err != nil {
		return false, err
	}
	return status.OK && status.DatabaseWritable && status.DataDir != "", nil
}

func (c *Client) ListKeys() ([]Key, error) {
	resp, err := c.http.Get(c.baseURL + "/api/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list keys failed: %s", resp.Status)
	}
	var keys []Key
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].Value != "" {
			keys[i].Value = "***"
		}
	}
	return keys, nil
}

func (c *Client) ListRoutes() ([]Route, error) {
	resp, err := c.http.Get(c.baseURL + "/api/routes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list routes failed: %s", resp.Status)
	}
	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, err
	}
	return routes, nil
}

func (c *Client) ListReferences() (*ReferenceReport, error) {
	resp, err := c.http.Get(c.baseURL + "/api/references")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list references failed: %s", resp.Status)
	}
	var report ReferenceReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, err
	}
	return &report, nil
}
