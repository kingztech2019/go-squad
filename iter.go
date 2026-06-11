package squad

import "context"

// Iter is a generic lazy iterator for paginated Squad API responses.
// Pages are fetched on demand — only when the current buffer is exhausted.
//
// Usage:
//
//	iter := client.Transfers.All(ctx, &squad.TransferListParams{PerPage: 50})
//	for iter.Next() {
//	    t := iter.Item()
//	    fmt.Println(t.TransactionRef, squad.FromKobo(t.Amount))
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
type Iter[T any] struct {
	ctx       context.Context
	fetchPage func(ctx context.Context, page int) ([]T, error)
	buf       []T
	cur       int
	page      int
	err       error
	exhausted bool
}

func newIter[T any](ctx context.Context, fetch func(context.Context, int) ([]T, error)) *Iter[T] {
	return &Iter[T]{
		ctx:       ctx,
		fetchPage: fetch,
		page:      1,
	}
}

// Next advances the iterator to the next item.
// Fetches the next page transparently when the current buffer is exhausted.
// Returns false when all items have been yielded or an error occurs.
func (i *Iter[T]) Next() bool {
	if i.err != nil || i.exhausted {
		return false
	}
	i.cur++
	if i.cur <= len(i.buf) {
		return true
	}
	// Buffer exhausted — fetch the next page.
	items, err := i.fetchPage(i.ctx, i.page)
	if err != nil {
		i.err = err
		return false
	}
	if len(items) == 0 {
		i.exhausted = true
		return false
	}
	i.buf = items
	i.cur = 1
	i.page++
	return true
}

// Item returns the current item. Must only be called after Next() returns true.
func (i *Iter[T]) Item() T {
	return i.buf[i.cur-1]
}

// Err returns the error that stopped iteration, or nil on clean completion.
func (i *Iter[T]) Err() error {
	return i.err
}
