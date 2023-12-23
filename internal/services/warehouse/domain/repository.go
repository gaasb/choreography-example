package domain

import (
	"context"
)

type WarehouseRepository interface {
	SaveInvoice(ctx context.Context, invoice Invoice) error
	UpdateInvoice(ctx context.Context, invoice Invoice) error
	GetInvoice(ctx context.Context, invoiceId string) (*Invoice, error)
	AddProduct(ctx context.Context, item Item) error
	GetProductBy(ctx context.Context, itemId int) (*Item, error)
	UpdateProduct(ctx context.Context, item Item) error
	DeleteProduct(ctx context.Context, itemId int) error
}
