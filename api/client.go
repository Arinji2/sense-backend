package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApiClient struct {
	BaseURL string
	Client  HTTPClient
}

func NewApiClient(baseURL string) *ApiClient {

	if baseURL == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file, in API Client Init")

		}
		baseURL = os.Getenv("PB_LINK")
	}
	return &ApiClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *ApiClient) doRequest(req *http.Request, authToken ...string) (map[string]interface{}, error) {

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

	result, err := c.doRequest(req, authToken...)

	if err != nil {
		return nil, fmt.Errorf("error from request doer %w", err)
	}

	return result, nil

}

func (c *ApiClient) SendRequestWithQuery(method, path string, query map[string]string, authToken ...string) (map[string]interface{}, error) {

	queryParams := url.Values{}
	for key, value := range query {
		queryParams.Add(key, value)
	}

	address, err := url.JoinPath(c.BaseURL, path)
	if err != nil {
		return nil, fmt.Errorf("error joining URL paths: %w", err)
	}

	fullURL := fmt.Sprintf("%s?%s", address, queryParams.Encode())
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	result, err := c.doRequest(req, authToken...)

	if err != nil {
		return nil, fmt.Errorf("error from request doer %w", err)
	}

	return result, nil

}
