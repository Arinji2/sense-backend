package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApiClient struct {
	BaseURL string
	Client  HTTPClient
}

func NewApiClient(baseURL string) *ApiClient {
	return &ApiClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *ApiClient) SendRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	jsonBody, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("error marshalling json %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil

}
