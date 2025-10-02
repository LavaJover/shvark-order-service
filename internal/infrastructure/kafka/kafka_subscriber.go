package publisher

import (
	"context"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type DefaultKafkaSubscriber struct {
    brokers []string
}

func NewDefaultKafkaSubscriber(brokers []string) *DefaultKafkaSubscriber {
    return &DefaultKafkaSubscriber{brokers: brokers}
}

func (k *DefaultKafkaSubscriber) Subscribe(topic, groupID string) (<-chan domain.Message, error) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: k.brokers,
        Topic:   topic,
        GroupID: groupID,
    })
    out := make(chan domain.Message)
    go func() {
        defer reader.Close()
        for {
            m, err := reader.ReadMessage(context.Background())
            if err != nil {
                close(out)
                return
            }
            out <- domain.Message{Key: m.Key, Value: m.Value}
        }
    }()
    return out, nil
}
