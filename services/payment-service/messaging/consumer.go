package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"payment-service/events"
	"payment-service/model"
	"payment-service/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderExchange    = "order_exchange"
	OrderRoutingKey  = "order.created"
	PaymentQueueName = "payment_order_created_queue"
)

type EventConsumer struct {
	url        string
	paymentSvc service.PaymentService
	stopChan   chan struct{}
}

func NewEventConsumer(url string, paymentSvc service.PaymentService) *EventConsumer {
	return &EventConsumer{
		url:        url,
		paymentSvc: paymentSvc,
		stopChan:   make(chan struct{}),
	}
}

func (c *EventConsumer) Start() {
	go c.run()
}

func (c *EventConsumer) Stop() {
	close(c.stopChan)
}

func (c *EventConsumer) run() {
	for {
		select {
		case <-c.stopChan:
			log.Println("[RABBITMQ] Consumer stopped")
			return
		default:
			log.Printf("[RABBITMQ] Connecting to RabbitMQ at %s...", c.url)
			if err := c.connectAndConsume(); err != nil {
				log.Printf("[RABBITMQ] Connection error: %v. Retrying in 3 seconds...", err)
				select {
				case <-c.stopChan:
					return
				case <-time.After(3 * time.Second):
					continue
				}
			}
		}
	}
}

func (c *EventConsumer) connectAndConsume() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 1. Declare durable exchange
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
		return err
	}

	// 2. Declare durable queue
	queue, err := ch.QueueDeclare(
		PaymentQueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}

	// 3. Bind queue to exchange with routing key
	err = ch.QueueBind(
		queue.Name,
		OrderRoutingKey,
		OrderExchange,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 4. Consume messages
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	log.Printf("[RABBITMQ] Consumer connected and listening on queue '%s'", queue.Name)

	closeChan := make(chan *amqp.Error)
	conn.NotifyClose(closeChan)

	for {
		select {
		case <-c.stopChan:
			return nil
		case amqpErr := <-closeChan:
			log.Printf("[RABBITMQ] Connection lost: %v", amqpErr)
			return amqpErr
		case d, ok := <-msgs:
			if !ok {
				log.Println("[RABBITMQ] Messages channel closed")
				return amqp.ErrClosed
			}
			c.handleDelivery(d)
		}
	}
}

func (c *EventConsumer) handleDelivery(d amqp.Delivery) {
	log.Println("[PAYMENT] OrderCreated received")

	var event events.OrderCreatedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshaling OrderCreated event: %v", err)
		_ = d.Nack(false, false)
		return
	}

	log.Printf("[PAYMENT] Processing payment for Order #%d (User: %d, Amount: $%.2f)", event.OrderID, event.UserID, event.Amount)

	req := model.ProcessPaymentRequest{
		OrderID:       event.OrderID,
		UserID:        event.UserID,
		Amount:        event.Amount,
		PaymentMethod: "RABBITMQ_ASYNC_SIMULATED",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payment, err := c.paymentSvc.ProcessPayment(ctx, req)
	if err != nil {
		log.Printf("[PAYMENT] Payment failed: %v", err)
		_ = d.Nack(false, true) // requeue
		return
	}

	log.Printf("[PAYMENT] Payment SUCCESS for Order #%d (PaymentID: %d, TxnID: %s)", payment.OrderID, payment.ID, payment.TransactionID)
	_ = d.Ack(false)
}
