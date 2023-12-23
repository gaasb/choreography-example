package domain

import (
	"errors"
)

var (
	ErrEqualToZero  = errors.New("qty equal zero")
	ErrLessThanZero = errors.New("less than zero")
)

type Quantity int

func (i Quantity) ToInt() int {
	return int(i)
}

func (i *Quantity) IncreaseBy(qty int) {
	*i += Quantity(qty)
}

func (i *Quantity) DecreaseBy(qty int) error {
	if i.ToInt() == 0 {
		return ErrEqualToZero
	}
	if i.ToInt()-qty <= 0 {
		return ErrLessThanZero
	}
	*i -= Quantity(qty)
	return nil
}
