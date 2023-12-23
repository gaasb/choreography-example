package domain

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryRepositoryImpl struct {
	sync.Mutex
	userBalances       map[string]Money
	userPaymentRecords map[string][]string
	payments           map[string]Payment
}

func NewInMemoryDB() *InMemoryRepositoryImpl {
	return &InMemoryRepositoryImpl{
		userBalances:       make(map[string]Money),
		userPaymentRecords: make(map[string][]string),
		payments:           make(map[string]Payment),
	}
}

func (r *InMemoryRepositoryImpl) SavePayment(ctx context.Context, payment Payment) error {
	if _, err := r.GetPayment(ctx, payment.aggregateId); err != nil {
		r.Lock()
		r.payments[payment.aggregateId] = payment
		r.RecordUserPayment(ctx, &payment)
		r.Unlock()
		return nil
	}
	return fmt.Errorf("payment with id: %s is already in database", payment.aggregateId)
}
func (r *InMemoryRepositoryImpl) UpdatePayment(ctx context.Context, payment Payment) error {
	if _, err := r.GetPayment(ctx, payment.aggregateId); err == nil {
		r.Lock()
		r.payments[payment.aggregateId] = payment
		r.Unlock()
		return nil
	}
	return fmt.Errorf("payment with id: %s not found", payment.aggregateId)
}
func (r *InMemoryRepositoryImpl) GetPayment(ctx context.Context, paymentId string) (*Payment, error) {
	if payment, ok := r.payments[paymentId]; ok {
		return &payment, nil
	}
	return nil, fmt.Errorf("payment with id: %s not found", paymentId)
}

func (r *InMemoryRepositoryImpl) GetBalance(ctx context.Context, userId string) (*Money, error) {
	if userBalance, ok := r.userBalances[userId]; ok {
		return &userBalance, nil
	}
	return nil, fmt.Errorf("user balance with id: %s not found", userId)
}

func (r *InMemoryRepositoryImpl) CreateBalance(ctx context.Context, userId string) error {
	if _, err := r.GetBalance(ctx, userId); err != nil {
		r.userBalances[userId] = NewMoney(0)
		return nil
	}
	return fmt.Errorf("user with id: %s is already taken", userId)
}

func (r *InMemoryRepositoryImpl) AddMoneyToBalance(ctx context.Context, userId string, amount Money) error {
	if userBalance, err := r.GetBalance(ctx, userId); err != nil {
		return err
	} else {
		r.Lock()
		userBalance.IncreaseBy(amount)
		r.UpdateBalance(ctx, userId, *userBalance)
		r.Unlock()
		return nil
	}
}
func (r *InMemoryRepositoryImpl) UpdateBalance(ctx context.Context, userId string, balance Money) {
	r.userBalances[userId] = balance
}
func (r *InMemoryRepositoryImpl) RecordUserPayment(ctx context.Context, payment *Payment) {
	r.userPaymentRecords[payment.customerId] = append(r.userPaymentRecords[payment.customerId], payment.aggregateId)
}
