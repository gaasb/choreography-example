package service

import (
	"choreography/internal/services/warehouse/domain"
	"encoding/json"

	"choreography/tools"
	"context"
	"fmt"
	"log/slog"
)

type WarehouseServiceImpl struct {
	db           domain.WarehouseRepository
	dispatcher   tools.EventDispatcher
	eventStorage *tools.EventStorage
	logger       *slog.Logger
}

type WarehouseRequest struct {
	OrderId    string      `json:"order_id"`
	CustomerId string      `json:"customer_id"`
	Line       domain.Line `json:"line"`
}

type WarehouseResponse struct {
	InvoiceId string `json:"invoice_id"`
}
type ProductRequest struct {
	ItemId int    `json:"item_id"`
	Name   string `json:"item_name"`
	Qty    int    `json:"qty"`
	Price  int    `json:"price"`
}

func (p ProductRequest) ToItem() domain.Item {
	return domain.NewItem(p.ItemId, p.Name, p.Qty, p.Price)
}

func NewWarehouseService(db domain.WarehouseRepository, dispatcher tools.EventDispatcher, storage *tools.EventStorage, logger *slog.Logger) WarehouseServiceImpl {
	return WarehouseServiceImpl{
		db:           db,
		dispatcher:   dispatcher,
		eventStorage: storage,
		logger:       logger,
	}
}
func NewWarehouseRequest(msg tools.EventMessage) (*WarehouseRequest, error) {

	orderId := msg.GetAggregateId()
	customerId := msg.Data["customer_id"].(string)
	var orderLine domain.Line
	if raw, err := json.Marshal(msg.Data["order_line"]); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(raw, &orderLine); err != nil {
			return nil, err
		}
	}
	return &WarehouseRequest{
		OrderId:    orderId,
		CustomerId: customerId,
		Line:       orderLine,
	}, nil
}

func (s *WarehouseServiceImpl) CreateInvoiceFromOrder(ctx context.Context, req *WarehouseRequest) (*WarehouseResponse, error) {

	var itemList []*domain.Item
	var err error

	if itemList, err = req.Line.ToItem(ctx, s.db); err != nil {
		slog.Warn("problem in converting line request to items", slog.String("err", err.Error()))
		return nil, err
	}
	for _, item := range itemList {
		qty := req.Line[item.GetItemId()]
		if err = item.DecreaseBy(domain.Quantity(qty)); err != nil {
			err = fmt.Errorf("decrease qty: %w", err)
		}
	}
	if err != nil {
		slog.Warn("problem with decrease quantity value in item", slog.String("err", err.Error()))
		return nil, err
	}
	for _, item := range itemList {
		if err = s.db.UpdateProduct(ctx, *item); err != nil {
			err = fmt.Errorf("update item: %w", err)
		}
	}
	if err != nil {
		slog.Warn("problem with updating item", slog.String("err", err.Error()))
		return nil, err
	}

	invoice := domain.NewInvoice(req.OrderId, req.CustomerId, itemList...)
	if err = s.db.SaveInvoice(ctx, invoice); err != nil {
		err = fmt.Errorf("invoice created but failed to insert into database: %w", err)
		slog.Warn("on save invoice", slog.String("err", err.Error()))
		return nil, err
	}
	slog.Info("invoice created", slog.String("id", invoice.GetId()))

	invoice.Process(ctx, s.dispatcher, s.eventStorage)
	return &WarehouseResponse{InvoiceId: invoice.GetId()}, nil
}

func (s *WarehouseServiceImpl) AddProduct(ctx context.Context, req ProductRequest) (bool, error) {
	item := req.ToItem()
	if err := s.db.AddProduct(ctx, item); err != nil {
		slog.Warn("on add product", slog.String("err", err.Error()))
		return false, err
	}
	slog.Info("added product with", slog.Int("id", req.ItemId))
	return true, nil
}

func (s *WarehouseServiceImpl) AcceptInvoice(ctx context.Context, invoiceId string) error {

	eventId := ctx.Value("event_id").(string)
	toService := ctx.Value("linked_service").(string)

	invoice, err := s.db.GetInvoice(ctx, invoiceId)
	if err != nil {
		slog.Warn("fail on accept invoice", slog.String("invoice_id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	invoice.Accept()

	restoredMessage := tools.NewEventMessage()
	restoredMessage.SetEventID(eventId)
	if err := restoredMessage.Restore(s.eventStorage); err != nil {
		slog.Warn("fail to restore invoice", slog.String("id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.SetState(invoice.GetStatus())
	restoredMessage.SetReceiver(toService)
	if err := invoice.Update(ctx, s.db); err != nil {
		slog.Warn("fail to update invoice", slog.String("id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.Save(s.eventStorage)
	slog.Info("invoice accepted", slog.String("invoice_id", invoiceId))
	s.dispatcher.Produce(ctx, restoredMessage)
	return nil
}
func (s *WarehouseServiceImpl) RejectInvoice(ctx context.Context, invoiceId string) error {

	eventId := ctx.Value("event_id").(string)

	invoice, err := s.db.GetInvoice(ctx, invoiceId)
	if err != nil {
		slog.Warn("fail on reject invoice", slog.String("invoice_id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	invoice.Reject()

	if err := invoice.RestoreItems(ctx, s.db); err != nil {
		slog.Warn("restore items in invoice", slog.String("err", err.Error()))
		return err
	}

	restoredMessage := tools.NewEventMessage()
	restoredMessage.SetEventID(eventId)
	if err := restoredMessage.Restore(s.eventStorage); err != nil {
		slog.Warn("fail to restore invoice", slog.String("id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.SetState(invoice.GetStatus())
	if v, ok := ctx.Value("linked_service").(string); ok {
		restoredMessage.SetReceiver(v)
	} else {
		restoredMessage.Header.Swap()
	}
	if err := invoice.Update(ctx, s.db); err != nil {
		slog.Warn("fail to update invoice", slog.String("id", invoiceId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.Save(s.eventStorage)
	slog.Info("invoice rejected and restored", slog.String("id", invoiceId))

	s.dispatcher.Produce(ctx, restoredMessage)

	return nil
}

func (s *WarehouseServiceImpl) StartDispatch(ctx context.Context, withEvents ...tools.EventOp) {
	slog.Info("Starting message dispatching", slog.String("service", s.GetServiceName()))
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Dispatcher stopped", slog.String("service", s.GetServiceName()))
				return
			default:
				s.dispatcher.Handle(ctx, withEvents...)
			}
		}

	}()
}
