package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"order-service/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderExchange   = "order_exchange"
	OrderRoutingKey = "order.created"
)

type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error
	Close()
}

type RabbitMQPublisher struct {
	url     string
	conn    *amqp.Connection
	ch      *amqp.Channel
	mu      sync.Mutex
	closing bool
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	pub := &RabbitMQPublisher{url: url}
	if err := pub.connect(); err != nil {
		return nil, err
	}
	return pub, nil
}

func (p *RabbitMQPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("failed to dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	// Declare durable exchange
	err = ch.ExchangeDeclare(
		OrderExchange, // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.conn = conn
	p.ch = ch
	return nil
}

func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
	p.mu.Lock()
	if p.ch == nil || p.conn == nil || p.conn.IsClosed() {
		p.mu.Unlock()
		if err := p.connect(); err != nil {
			return fmt.Errorf("rabbitmq reconnect error: %w", err)
		}
		p.mu.Lock()
	}
	ch := p.ch
	p.mu.Unlock()

	log.Printf("[ORDER] Publishing OrderCreated for Order #%d (Amount: $%.2f)", event.OrderID, event.Amount)

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order event: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		pubCtx,
		OrderExchange,
		OrderRoutingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Persistent delivery mode (2)
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to rabbitmq: %w", err)
	}

	log.Printf("[RABBITMQ] OrderCreated published (RoutingKey: %s, OrderID: %d)", OrderRoutingKey, event.OrderID)
	return nil
}

func (p *RabbitMQPublisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closing = true
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

// Mock Publisher for unit tests or when RabbitMQ is offline
type MockEventPublisher struct {
	mu              sync.Mutex
	PublishedEvents []events.OrderCreatedEvent
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		PublishedEvents: []events.OrderCreatedEvent{},
	}
}

func (m *MockEventPublisher) PublishOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.Printf("[ORDER] Publishing OrderCreated (Mock)")
	m.PublishedEvents = append(m.PublishedEvents, event)
	log.Printf("[RABBITMQ] OrderCreated published (Mock)")
	return nil
}

func (m *MockEventPublisher) Close() {}
