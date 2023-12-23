package domain

import "errors"

type Line map[int]int

func NewItemLine() Line {
	return make(Line)
}

func (i Line) AddItem(itemId int, qty int) error {
	if qty <= 0 {
		return errors.New("invalid quantity of items")
	}
	i[itemId] = qty
	return nil
}

func (i Line) Verify() error {
	for _, v := range i {
		if v <= 0 {
			return errors.New("invalid quantity of items")
		}
	}
	return nil
}
