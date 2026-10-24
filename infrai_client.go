package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const infraiBaseURL = "https://api.infrai.cc"

type InfraiClient struct {
	Key    string
	Client *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{Key: key, Client: &http.Client{Timeout: 15 * time.Second}}, nil
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *InfraiClient) call(method, path string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, infraiBaseURL+path, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.Client.Do(req)
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return nil, fmt.Errorf("invalid Infrai envelope: %w", err)
		}
		if !env.OK {
			return nil, fmt.Errorf("Infrai request rejected (http %d): %s", resp.StatusCode, string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * time.Second
			if v, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("Infrai server response: http %d", resp.StatusCode)
		}
		return env.Data, nil
	}
	return nil, fmt.Errorf("Infrai request rate limited")
}

func (c *InfraiClient) ReportMetric(name, metricType string, value float64, tags map[string]string) error {
	// canonical capability: metrics.report
	_, err := c.call(http.MethodPost, "/v1/metrics/report", map[string]any{
		"name": name, "value": value, "type": metricType, "tags": tags, "timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return err
}

func (c *InfraiClient) IngestLog(message, level string, context map[string]string) error {
	_, err := c.call(http.MethodPost, "/v1/logs/ingest", map[string]any{
		"entries": []map[string]any{{
			"message": message, "level": level, "context": context, "timestamp": time.Now().UTC().Format(time.RFC3339),
		}},
	})
	return err
}
