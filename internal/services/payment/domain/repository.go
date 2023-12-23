package domain

import (
	"context"
)

type PaymentRepository interface {
	SavePayment(ctx context.Context, payment Payment) error
	GetPayment(ctx context.Context, paymentId string) (*Payment, error)
	UpdatePayment(ctx context.Context, payment Payment) error
	CreateBalance(ctx context.Context, userId string) error
	GetBalance(ctx context.Context, userId string) (*Money, error)
	AddMoneyToBalance(ctx context.Context, userId string, amount Money) error
	UpdateBalance(ctx context.Context, userId string, balance Money)
	RecordUserPayment(ctx context.Context, payment *Payment)
}
