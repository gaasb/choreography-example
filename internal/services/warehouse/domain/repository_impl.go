package domain

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryRepositoryImpl struct {
	sync.Mutex
	items    map[int]Item
	invoices map[string]Invoice
}

func NewInMemoryDB() *InMemoryRepositoryImpl {
	return &InMemoryRepositoryImpl{
		items:    make(map[int]Item),
		invoices: make(map[string]Invoice),
	}
}

func (r *InMemoryRepositoryImpl) SaveInvoice(ctx context.Context, invoice Invoice) error {
	if _, err := r.GetInvoice(ctx, invoice.aggregateId); err != nil {
		r.Lock()
		r.invoices[invoice.aggregateId] = invoice
		r.Unlock()
		return nil
	}
	return fmt.Errorf("invoice with id: %s is already in database", invoice.aggregateId)
}
func (r *InMemoryRepositoryImpl) UpdateInvoice(ctx context.Context, invoice Invoice) error {
	if _, err := r.GetInvoice(ctx, invoice.aggregateId); err == nil {
		r.Lock()
		r.invoices[invoice.aggregateId] = invoice
		r.Unlock()
		return nil
	}
	return fmt.Errorf("invoice with id: %s not found", invoice.aggregateId)
}
func (r *InMemoryRepositoryImpl) GetInvoice(ctx context.Context, invoiceId string) (*Invoice, error) {
	if invocie, ok := r.invoices[invoiceId]; ok {
		return &invocie, nil
	}
	return nil, fmt.Errorf("invoice with id: %s not found", invoiceId)
}

func (r *InMemoryRepositoryImpl) AddProduct(ctx context.Context, item Item) error {
	if item.itemId == 0 {
		item.itemId = len(r.items) + 1
	}
	_, err := r.GetProductBy(ctx, item.itemId)
	if err != nil {
		r.Lock()
		r.items[item.itemId] = item
		defer r.Unlock()
		return nil
	}
	return fmt.Errorf("item wit id: %d already in database", item.itemId)
}
func (r *InMemoryRepositoryImpl) GetProductBy(ctx context.Context, itemId int) (*Item, error) {
	if item, ok := r.items[itemId]; ok {
		return &item, nil
	}
	return nil, fmt.Errorf("item with id: %d not found", itemId)
}
func (r *InMemoryRepositoryImpl) UpdateProduct(ctx context.Context, item Item) error {
	if _, err := r.GetProductBy(ctx, item.itemId); err != nil {
		return err
	}
	r.Lock()
	r.items[item.itemId] = item
	r.Unlock()
	return nil
}
func (r *InMemoryRepositoryImpl) DeleteProduct(ctx context.Context, itemId int) error {
	if _, err := r.GetProductBy(ctx, itemId); err != nil {
		return err
	}
	r.Lock()
	delete(r.items, itemId)
	r.Unlock()
	return nil
}
