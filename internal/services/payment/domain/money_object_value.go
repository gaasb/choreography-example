package domain

import (
	"fmt"
)

type Money int

func NewMoney(value int) Money {
	if value < 0 {
		return Money(0)
	}
	return Money(value)
}

func (m Money) ToInt() int {
	return int(m)
}
func (m *Money) IncreaseBy(amount Money) {
	*m += amount
}

func (m *Money) DecreaseBy(amount Money) error {
	if *m-amount < 0 {
		return fmt.Errorf("not enough money")
	}
	*m -= amount
	return nil
}
