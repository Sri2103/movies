package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"movieexample.com/rating/pkg/model"
)

const kafkaTImeOut = 10 * time.Second

func main() {
	fmt.Println("Creating a kafka producer...")

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	const fileName = "ratingsdara.json"

	fmt.Println("Reading data from file: ", fileName)

	ratings, err := readRatingsFromFile(fileName)
	if err != nil {
		panic(err)
	}

	err = produceRatingEvents("ratings", producer, ratings)
	if err != nil {
		panic(err)
	}

	timeout := kafkaTImeOut

	fmt.Println("Waiting for all messages to be delivered...")
	producer.Flush(int(timeout.Milliseconds()))
}

func readRatingsFromFile(fileName string) ([]model.RatingEvent, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	var ratings []model.RatingEvent
	if err := json.NewDecoder(f).Decode(&ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}

func produceRatingEvents(topic string, producer *kafka.Producer, events []model.RatingEvent) error {
	for _, event := range events {
		encodedEvent, err := json.Marshal(event)
		if err != nil {
			return err
		}

		if err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: int32(kafka.PartitionAny)},
			Value:          []byte(encodedEvent),
		}, nil); err != nil {
			return err
		}
	}

	return nil
}
