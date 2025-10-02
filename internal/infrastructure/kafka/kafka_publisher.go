package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type DefaultKafkaPublisher struct {
	writer *kafka.Writer
}

func NewDefaultKafkaPublisher(brokers []string) *DefaultKafkaPublisher {
	return &DefaultKafkaPublisher{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *DefaultKafkaPublisher) Publish(topic string, msgs ...domain.Message) error {
	var km []kafka.Message
	for _, m := range msgs {
		km = append(km, kafka.Message{
			Key: m.Key,
			Value: m.Value,
			Time: time.Now(),
			Topic: topic,
		})
	}

	return k.writer.WriteMessages(context.Background(), km...)
}

func (k *DefaultKafkaPublisher) PublishOrder(topic string, event OrderEvent) error {
	v, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return k.Publish(topic, domain.Message{Key: []byte(event.TraderID), Value: v})
}

// BatchPublishOrders - батчевая публикация событий заказов
func (k *DefaultKafkaPublisher) BatchPublishOrders(topic string, events []OrderEvent) error {
    if len(events) == 0 {
        return nil
    }

    // Если только одно событие - используем обычную публикацию
    if len(events) == 1 {
        return k.PublishOrder(topic, events[0])
    }

    // Подготавливаем batch сообщений
    messages := make([]kafka.Message, 0, len(events))
    timestamp := time.Now()

    for _, event := range events {
        msg, err := json.Marshal(event)
        if err != nil {
            log.Printf("Failed to marshal event for order %s: %v", event.OrderID, err)
            continue // Пропускаем проблемное сообщение, но продолжаем с остальными
        }

        messages = append(messages, kafka.Message{
            Key:   []byte(event.TraderID),
            Value: msg,
            Time:  timestamp,
        })
    }

    if len(messages) == 0 {
        return fmt.Errorf("no valid messages to publish")
    }

    // Публикуем batch с контекстом и таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := k.writer.WriteMessages(ctx, messages...); err != nil {
        return fmt.Errorf("failed to write batch messages: %w", err)
    }

    log.Printf("Successfully published %d order events to Kafka", len(messages))
    return nil
}

// BatchPublishOrdersWithRetry - батчевая публикация с retry и разбивкой на части
func (k *DefaultKafkaPublisher) BatchPublishOrdersWithRetry(topic string, events []OrderEvent, batchSize int, maxRetries int) error {
    if len(events) == 0 {
        return nil
    }

    // Разбиваем на батчи если слишком много событий
    if batchSize <= 0 {
        batchSize = 100 // По умолчанию 100 сообщений в батче
    }

    var allErrors []error
    successfulCount := 0

    for i := 0; i < len(events); i += batchSize {
        end := i + batchSize
        if end > len(events) {
            end = len(events)
        }

        batch := events[i:end]
        
        // Пытаемся опубликовать батч с retry
        var err error
        for attempt := 1; attempt <= maxRetries; attempt++ {
            err = k.BatchPublishOrders(topic, batch)
            if err == nil {
                successfulCount += len(batch)
                break
            }

            log.Printf("Batch publish attempt %d failed: %v", attempt, err)
            
            // Экспоненциальная задержка между попытками
            if attempt < maxRetries {
                time.Sleep(time.Duration(attempt) * time.Second)
            }
        }

        if err != nil {
            allErrors = append(allErrors, fmt.Errorf("batch %d-%d failed after %d attempts: %w", 
                i, end, maxRetries, err))
        }
    }

    log.Printf("Batch publish completed: %d/%d events successful", successfulCount, len(events))

    // Возвращаем ошибку только если все батчи провалились
    if successfulCount == 0 && len(allErrors) > 0 {
        return fmt.Errorf("all batches failed: %v", allErrors)
    }

    return nil
}