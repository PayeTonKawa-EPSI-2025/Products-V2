package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishProductEvent publishes a product event to RabbitMQ
func PublishProductEvent(ch *amqp.Channel, eventType events.EventType, product models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := events.ProductEvent{
		Type:      eventType,
		Product:   product,
		Timestamp: time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return err
	}

	// Use a routing key based on the event type
	routingKey := string(eventType)

	err = ch.PublishWithContext(
		ctx,
		"events", // exchange
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("Published %s event for product %d", eventType, product.ID)
	return nil
}
