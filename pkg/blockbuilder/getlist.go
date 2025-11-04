package blockbuilder

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
)

var ErrBadStatusCode = errors.New("bad HTTP status code")

func getList(ctx context.Context, client *http.Client, url string) (results []string, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		return nil, fmt.Errorf("%w: %d %s", ErrBadStatusCode, response.StatusCode, response.Status)
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			results = append(results, line)
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("scanning: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}
