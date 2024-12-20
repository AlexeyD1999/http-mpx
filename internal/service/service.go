package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"mpx/internal/models"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	maxCountInputUrls = 20
	maxParallelReqs   = 4
	timeoutDuration   = 1 * time.Second
)

var errMaxCountReached = errors.New("max count of URLs reached")

type Service struct {
	client *http.Client
}

func NewService() *Service {
	client := &http.Client{
		Timeout: timeoutDuration,
	}

	return &Service{client: client}
}

// Send processes the URLs with constraints on concurrency and timeout
func (s *Service) Send(ctx context.Context, urls []string) (*models.Response, error) {
	// Check if the number of URLs exceeds the max allowed
	if len(urls) > maxCountInputUrls {
		return nil, errMaxCountReached
	}

	var wg sync.WaitGroup
	// Channel to limit parallelism to maxParallelReqs
	semaphore := make(chan struct{}, maxParallelReqs)

	// Error channel to capture errors from goroutines
	errCh := make(chan error, maxParallelReqs)

	var results []models.Result
	var mu sync.Mutex

	for _, inputURL := range urls {
		semaphore <- struct{}{}

		wg.Add(1)

		go func(inputURL string) {
			defer wg.Done()

			response, err := s.sendRequest(ctx, inputURL)
			if err != nil {
				errCh <- fmt.Errorf("failed to send request to url: %s: %w", inputURL, err)

				<-semaphore

				return
			}

			var result models.Result

			if err = json.Unmarshal(response, &result); err != nil {
				errCh <- fmt.Errorf("failed to unmarshal response: %w", err)

				<-semaphore

				return
			}

			mu.Lock()
			result.URL = inputURL
			results = append(results, result)
			mu.Unlock()

			<-semaphore
		}(inputURL)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return nil, fmt.Errorf("failed to send request to urls: %w", err)
	}

	return &models.Response{Results: results}, nil
}

func (s *Service) sendRequest(ctx context.Context, inputURL string) ([]byte, error) {
	if _, err := url.ParseRequestURI(inputURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, inputURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	response, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer response.Body.Close()

	resBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response bytes: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unknown error in response body")
	}

	return resBytes, nil
}
