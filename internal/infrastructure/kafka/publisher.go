package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaPublisher) PublishDispute(event DisputeEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.TraderID),
		Value: msg,
		Time:  time.Now(),
	})
}

func (k *KafkaPublisher) Publish(event OrderEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.TraderID),
		Value: msg,
		Time:  time.Now(),
	})
}
