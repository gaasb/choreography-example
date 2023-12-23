package domain

import (
	"context"
	"errors"
	"fmt"
)

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

func (s Line) ToItem(ctx context.Context, repository WarehouseRepository) ([]*Item, error) {
	var errOut error
	output := []*Item{}
	for itemId := range s {
		if item, err := repository.GetProductBy(ctx, itemId); err != nil {
			errOut = fmt.Errorf("converting line to item: %w", err)
		} else if errOut == nil {
			output = append(output, item)
		}
	}
	if errOut != nil {
		return nil, errOut
	}
	return output, nil
}
