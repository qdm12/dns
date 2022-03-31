package blockbuilder

import (
	"context"
	"net/http"
)

func getLists(ctx context.Context, client *http.Client, urls []string) (
	uniqueResults map[string]struct{}, errs []error) {
	chResults := make(chan []string)
	chError := make(chan error)
	for _, url := range urls {
		go func(url string) {
			results, err := getList(ctx, client, url)
			chResults <- results
			chError <- err
		}(url)
	}

	uniqueResults = make(map[string]struct{})
	listsLeftToFetch := len(urls)
	for listsLeftToFetch > 0 {
		select {
		case results := <-chResults:
			for _, result := range results {
				uniqueResults[result] = struct{}{}
			}
		case err := <-chError:
			listsLeftToFetch--
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return uniqueResults, errs
}
