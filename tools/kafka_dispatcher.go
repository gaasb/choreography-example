package tools

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	ip   string
	port int
)

type EventDispatcherKafka struct {
	broker *kgo.Client
}

func init() {
	kafkaAddr()
}

func NewEventDispatcherKafka(serviceName string) *EventDispatcherKafka {
	connAddr := fmt.Sprintf("%s:%d", ip, port)
	client, err := kgo.NewClient(
		kgo.SeedBrokers(connAddr),
		kgo.ConsumeTopics(serviceName),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	err = client.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	_, err = kadm.NewClient(client).CreateTopic(context.Background(), 1, 1, nil, serviceName)
	if err != nil {
		fmt.Println(err.Error())
	}
	dispatcher := &EventDispatcherKafka{
		broker: client,
	}
	return dispatcher
}

func (d *EventDispatcherKafka) Handle(ctx context.Context, operation ...EventOp) {
	currentService := ctx.Value("current_service").(string)

	select {
	case <-ctx.Done():
		d.broker.Close()
		return
	default:
		fetches := d.broker.PollFetches(ctx)
		iter := fetches.RecordIter()
		for !iter.Done() {
			rawMsg := iter.Next()
			var message EventMessage
			if err := json.Unmarshal(rawMsg.Value, &message); err != nil {
				continue
			}
			for _, run := range operation {
				run(message)
			}
			slog.Info(
				"event handled",
				slog.String("event_id", message.GetEventID()),
				slog.String("event_type", message.GetEventName()),
				slog.String("from", message.GetProducer()),
				slog.String("to", message.GetReceiver()),
				slog.String("service", currentService),
			)
		}
	}
}

func (d *EventDispatcherKafka) Produce(ctx context.Context, message EventMessage) {
	currentService := ctx.Value("current_service").(string)
	var err error
	var rawMsg []byte

	if rawMsg, err = json.Marshal(message); err != nil {
		slog.Error("dispatcher on marshal message", slog.String("err", err.Error()))
		return
	}
	msg := kgo.Record{Topic: message.GetReceiver(), Value: rawMsg}
	d.broker.Produce(ctx, &msg, func(_ *kgo.Record, errOut error) {
		err = errOut
	})
	if err != nil {
		slog.Error("dispatcher on produce message", slog.String("err", err.Error()))
		return
	}
	slog.Info(
		"event produced",
		slog.String("event_id", message.GetEventID()),
		slog.String("event_type", message.GetEventName()),
		slog.String("from", message.GetProducer()),
		slog.String("to", message.GetReceiver()),
		slog.String("service", currentService),
	)
}

func kafkaAddr() {
	flag.StringVar(&ip, "ip", "localhost", "ip address for kafka connection")
	flag.IntVar(&port, "port", 9092, "port for kafka connection")
	flag.Parse()
}
