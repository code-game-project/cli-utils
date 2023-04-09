package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func FetchFile(url string) (io.ReadCloser, error) {
	// TODO: implement caching
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch '%s': %w", url, err)
	}
	return resp.Body, nil
}

func FetchJSON[T any](url string) (T, error) {
	var obj T
	file, err := FetchFile(url)
	if err != nil {
		return obj, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("failed to decode response from '%s': %w", url, err)
	}
	return obj, nil
}
