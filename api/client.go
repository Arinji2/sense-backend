package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func (c *ApiClient) SendRequestWithBody(method, path string, body interface{}, authToken ...string) (map[string]interface{}, error) {
	address := fmt.Sprintf("%s%s", c.BaseURL, path)
	jsonBody, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("error marshalling json %w", err)
	}

	req, err := http.NewRequest(method, address, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(authToken) != 0 && authToken[0] != "" {
		req.Header.Set("Authorization", authToken[0])
	}
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

func (c *ApiClient) SendRequestWithQuery(method, path string, query map[string]string, authToken ...string) (map[string]interface{}, error) {

	var queryParams strings.Builder
	for key, value := range query {
		parsedKey := url.QueryEscape(key)
		parsedValue := url.QueryEscape(value)
		if queryParams.Len() > 0 {
			queryParams.WriteString("&")
		}
		queryParams.WriteString(parsedKey)
		queryParams.WriteString("=")
		queryParams.WriteString(parsedValue)
	}

	address := fmt.Sprintf("%s%s?=%s", c.BaseURL, path, &queryParams)

	req, err := http.NewRequest(method, address, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(authToken) != 0 && authToken[0] != "" {
		req.Header.Set("Authorization", authToken[0])
	}
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
