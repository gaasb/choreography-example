package domain

import (
	"context"
)

type OrderRepository interface {
	Add(ctx context.Context, order Order) error
	GetBy(ctx context.Context, id string) (*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
}
