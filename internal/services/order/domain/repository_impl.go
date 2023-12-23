package domain

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryRepositoryImpl struct {
	sync.Mutex
	data map[string]Order
}

func NewInMemoryDB() *InMemoryRepositoryImpl {
	return &InMemoryRepositoryImpl{
		data: make(map[string]Order),
	}
}

func (r *InMemoryRepositoryImpl) Add(ctx context.Context, order Order) error {
	if _, err := r.GetBy(ctx, order.aggregateId); err != nil {
		r.Lock()
		r.data[order.aggregateId] = order
		r.Unlock()
		return nil
	}
	return fmt.Errorf("order with id: %s already in database", order.aggregateId)
}
func (r *InMemoryRepositoryImpl) GetBy(ctx context.Context, id string) (*Order, error) {
	if order, ok := r.data[id]; ok {
		return &order, nil
	}
	return nil, fmt.Errorf("order with id: %s not found", id)
}
func (r *InMemoryRepositoryImpl) Update(ctx context.Context, order *Order) error {
	if _, err := r.GetBy(ctx, order.aggregateId); err != nil {
		return err
	}
	r.Lock()
	r.data[order.aggregateId] = *order
	r.Unlock()
	return nil
}
func (r *InMemoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	if _, err := r.GetBy(ctx, id); err != nil {
		return err
	}
	r.Lock()
	delete(r.data, id)
	r.Unlock()
	return nil
}
