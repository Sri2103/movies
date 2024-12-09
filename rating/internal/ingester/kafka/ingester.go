package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"movieexample.com/rating/pkg/model"
)

// Ingester is a struct that represents a Kafka ingester. It contains a Kafka consumer
// and the topic to consume from.
type Ingester struct {
	consumer *kafka.Consumer
	topic    string
}

// NewIngester creates a new Kafka ingester. It takes the Kafka broker address,
// the consumer group ID, and the topic to consume from as arguments. It returns
// a new Ingester instance and any error that occurred during initialization.
func NewIngester(addr string, groupID string, topic string) (*Ingester, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	return &Ingester{consumer: consumer, topic: topic}, nil
}

func (i *Ingester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {
	if err := i.consumer.SubscribeTopics([]string{i.topic}, nil); err != nil {
		return nil, err
	}
	ch := make(chan model.RatingEvent)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				i.consumer.Close()
				return
			default:
				msg, err := i.consumer.ReadMessage(-1)
				if err != nil {
					continue
				}
				var event model.RatingEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					fmt.Println("Unmarshal error: " + err.Error())
					continue
				}
				ch <- event
			}
		}
	}()
	return ch, nil
}
