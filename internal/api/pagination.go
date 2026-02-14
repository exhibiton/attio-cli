package api

import "context"

// FetchAllOffset collects all pages from an offset-paginated endpoint.
func FetchAllOffset[T any](ctx context.Context, limit int, maxPages int, fetch func(offset int) ([]T, error)) ([]T, error) {
	if limit <= 0 {
		limit = 1
	}
	if maxPages <= 0 {
		maxPages = 100
	}

	all := make([]T, 0)
	offset := 0
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return all, err
		}
		items, err := fetch(offset)
		if err != nil {
			return all, err
		}
		all = append(all, items...)
		if len(items) < limit {
			break
		}
		offset += len(items)
	}
	return all, nil
}

// FetchAllCursor collects all pages from a cursor-paginated endpoint.
func FetchAllCursor[T any](ctx context.Context, maxPages int, fetch func(cursor string) ([]T, string, error)) ([]T, error) {
	if maxPages <= 0 {
		maxPages = 100
	}

	all := make([]T, 0)
	cursor := ""
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return all, err
		}
		items, nextCursor, err := fetch(cursor)
		if err != nil {
			return all, err
		}
		all = append(all, items...)
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return all, nil
}
