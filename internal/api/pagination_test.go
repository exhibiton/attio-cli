package api

import (
	"context"
	"errors"
	"testing"
)

func TestFetchAllOffset(t *testing.T) {
	t.Parallel()

	pages := map[int][]int{
		0: {1, 2},
		2: {3, 4},
		4: {5},
	}

	items, err := FetchAllOffset(context.Background(), 2, 10, func(offset int) ([]int, error) {
		return pages[offset], nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
}

func TestFetchAllOffsetStopsOnError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	items, err := FetchAllOffset(context.Background(), 2, 10, func(offset int) ([]int, error) {
		if offset >= 2 {
			return nil, wantErr
		}
		return []int{1, 2}, nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(items) != 2 {
		t.Fatalf("expected partial items before error, got %d", len(items))
	}
}

func TestFetchAllCursor(t *testing.T) {
	t.Parallel()

	items, err := FetchAllCursor(context.Background(), 10, func(cursor string) ([]int, string, error) {
		switch cursor {
		case "":
			return []int{1, 2}, "c1", nil
		case "c1":
			return []int{3}, "", nil
		default:
			return nil, "", nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestFetchAllOffsetContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	items, err := FetchAllOffset(ctx, 2, 10, func(_ int) ([]int, error) {
		called = true
		return []int{1, 2}, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if called {
		t.Fatalf("fetch callback should not be called once context is canceled")
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %d", len(items))
	}
}

func TestFetchAllCursorContextCanceledAfterFirstPage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	items, err := FetchAllCursor(ctx, 10, func(cursor string) ([]int, string, error) {
		calls++
		if cursor == "" {
			cancel()
			return []int{1}, "next", nil
		}
		return []int{2}, "", nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected a single call before cancellation, got %d", calls)
	}
	if len(items) != 1 || items[0] != 1 {
		t.Fatalf("expected partial first page, got %#v", items)
	}
}
