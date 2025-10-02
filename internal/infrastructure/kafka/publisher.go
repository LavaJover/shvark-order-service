package publisher

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaConfig struct {
	Brokers    []string
    Topic      string
    Username   string
    Password   string
    Mechanism  string // "PLAIN", "SCRAM-SHA-256", etc.
    TLSEnabled bool
}

type KafkaPublisher struct {
    writer *kafka.Writer
}

func NewKafkaPublisher(cfg KafkaConfig) (*KafkaPublisher, error) {
    // Базовые настройки writer
    writerConfig := kafka.Writer{
        Addr:     kafka.TCP(cfg.Brokers...),
        Topic:    cfg.Topic,
        Balancer: &kafka.LeastBytes{},
    }

    // Если включена аутентификация SASL
    if cfg.Username != "" && cfg.Password != "" {
        mechanism, err := createSASLMechanism(cfg.Mechanism, cfg.Username, cfg.Password)
        if err != nil {
            return nil, err
        }

        // Настраиваем транспорт с SASL и TLS
        writerConfig.Transport = &kafka.Transport{
            SASL: mechanism,
            TLS:  createTLSConfig(cfg.TLSEnabled),
        }
    }

    return &KafkaPublisher{
        writer: &writerConfig,
    }, nil
}

func createSASLMechanism(mechanism, username, password string) (sasl.Mechanism, error) {
    switch mechanism {
    case "SCRAM-SHA-256":
        return scram.Mechanism(scram.SHA256, username, password)
    case "SCRAM-SHA-512":
        return scram.Mechanism(scram.SHA512, username, password)
    case "PLAIN":
        return plain.Mechanism{
            Username: username,
            Password: password,
        }, nil
    default:
        return nil, fmt.Errorf("unsupported SASL mechanism: %s", mechanism)
    }
}

func createTLSConfig(enabled bool) *tls.Config {
    if !enabled {
        return nil
    }
    return &tls.Config{
        InsecureSkipVerify: false, // В продакшене должно быть false
    }
}

func (k *KafkaPublisher) PublishOrder(event OrderEvent) error {
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