// Package usersummary provides an example of a core business API that
// is based on a view.
package usersummary

import (
	"context"
	"fmt"

	"github.com/aleury/service/business/data/order"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Summary, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Core manages the set of API's for user summaries.
type Core struct {
	storer Storer
}

// NewCore constructs a Core for user summary api access.
func NewCore(storer Storer) *Core {
	return &Core{
		storer: storer,
	}
}

// Query retrieves a list of usersummary based on the provided filter.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Summary, error) {
	summaries, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return summaries, nil
}

// Count returns the number of usersummary records in the database.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := c.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}
