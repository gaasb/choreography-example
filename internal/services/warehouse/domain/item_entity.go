package domain

import (
	"errors"
	"fmt"
)

type Item struct {
	itemId int
	name   string
	qty    Quantity
	price  Money
}

func NewItem(id int, name string, quantity int, price int) Item {
	item := Item{
		itemId: id,
		name:   name,
		qty:    Quantity(quantity),
		price:  Money(price),
	}
	return item
}

func (i *Item) GetItemId() int {
	return i.itemId
}

func (i *Item) IncreaseBy(qty Quantity) {
	i.qty.IncreaseBy(qty.ToInt())
}

func (i *Item) DecreaseBy(qty Quantity) error {
	err := i.qty.DecreaseBy(qty.ToInt())

	if errors.Is(err, ErrEqualToZero) {
		return fmt.Errorf("item with id: %d out of stock", i.itemId)
	}
	if errors.Is(err, ErrLessThanZero) {
		return fmt.Errorf("not enough items with id: %d in stock", i.itemId)
	}
	return nil
}
