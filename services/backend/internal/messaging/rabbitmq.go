package messaging

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type RMQClient struct {
	url      string
	conn     *amqp.Connection
	ch       *amqp.Channel
	mu       sync.RWMutex
	isClosed bool
}

type AMQPHeaderCarrier amqp.Table

type MessageHandler func(ctx context.Context, body []byte) error

// NewRMQClient initializes the client structure configuration
func NewRMQClient(user, password, host string, port int, vhost string) *RMQClient {
	cleanVhost := strings.TrimPrefix(vhost, "/")

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, password, host, port, cleanVhost)
	slog.Info("new RMQClient instance", "host", host, "port", port, "vhost", vhost)
	return &RMQClient{url: url}
}

// Connect establishes the network socket and handles recovery loops
func (r *RMQClient) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("RMQClient connecting...")
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to dial rabbitmq broker: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open messaging channel: %w", err)
	}

	r.conn = conn
	r.ch = ch
	r.isClosed = false

	// Spin up a background listener to handle unexpected connection drops
	go r.handleDisconnect()

	return nil
}

// handleDisconnect monitors the connection closure channel and triggers a reconnect loop
func (r *RMQClient) handleDisconnect() {
	closeChan := r.conn.NotifyClose(make(chan *amqp.Error))

	err := <-closeChan
	if err != nil {
		log.Printf("[RMQ] Connection lost: %v. Initiating reconnection cycle...", err)

		r.mu.Lock()
		if r.isClosed {
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()

		// backoff reconnection loop until the broker is recovered
		for {
			time.Sleep(3 * time.Second)
			log.Println("[RMQ] Attempting to reconnect to broker box...")
			if err := r.Connect(); err == nil {
				log.Println("[RMQ] Reconnection successfully established!")
				return
			}
		}
	}
}

// Publish dispatches a byte payload to a specific target exchange or queue route
func (r *RMQClient) Publish(ctx context.Context, exchangeName, routingKey string, body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	headers := amqp.Table{}
	otel.GetTextMapPropagator().Inject(ctx, AMQPHeaderCarrier(headers))

	if r.ch == nil || r.conn.IsClosed() {
		slog.ErrorContext(ctx, "failed to publish message", "routing_key", routingKey, "reason", "client disconnected")
		return fmt.Errorf("cannot publish message: rabbitmq client is disconnected")
	}

	err := r.ch.PublishWithContext(ctx,
		exchangeName, // exchange name string
		routingKey,   // target queue routing identifier
		false,        // mandatory flag
		false,        // immediate flag
		amqp.Publishing{
			ContentType:  "application/json",
			Headers:      headers,
			DeliveryMode: amqp.Persistent, // Forces message persistence across broker restarts
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to publish message", "routing_key", routingKey, "error", err)
		return fmt.Errorf("message publication failed: %w", err)
	}

	return nil
}

func (r *RMQClient) Consume(ctx context.Context, queueName, consumer string) (<-chan amqp.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.ch == nil || r.conn.IsClosed() {
		slog.ErrorContext(ctx, "failed to consume queue", "queue_name", queueName, "reason", "client disconnected")
		return nil, fmt.Errorf("cannot hook into queue")
	}

	deliveries, err := r.ch.ConsumeWithContext(ctx,
		queueName,
		consumer,
		false, // autoAck flag
		false, // exclusive flag
		false, // noLocal flag
		false, // noWait flag
		amqp.Table{},
	)

	if err != nil {
		return nil, fmt.Errorf("consumption of the queue %s failed: %w", queueName, err)
	}

	return deliveries, nil
}

func (r *RMQClient) Ack(tag uint64, multiple bool) error {
	return r.ch.Ack(tag, multiple)
}

func (r *RMQClient) Nack(tag uint64, multiple bool, requeue bool) error {
	return r.ch.Nack(tag, multiple, requeue)
}

func (r *RMQClient) StartConsumer(
	ctx context.Context,
	wg *sync.WaitGroup,
	queueName string,
	consumerTag string,
	handler MessageHandler,
) error {
	deliveries, err := r.Consume(ctx, queueName, consumerTag)
	if err != nil {
		return err
	}

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					slog.WarnContext(ctx, "rabbitmq deliveries channel closed", "queue", queueName)
					return
				}
				r.processDelivery(ctx, queueName, d, handler)
			}
		}
	})

	return nil
}

func (r *RMQClient) processDelivery(ctx context.Context, queueName string, d amqp.Delivery, handler MessageHandler) {
	msgCtx, span := ExtractDeliveryContext(ctx, d, "process-"+queueName)
	defer span.End()

	if len(d.Body) == 0 {
		slog.WarnContext(msgCtx, "received empty rabbitmq message, discarding", "queue", queueName)
		_ = r.Nack(d.DeliveryTag, false, false) // false = do not requeue poison/empty message
		return
	}

	if err := handler(msgCtx, d.Body); err != nil {
		slog.ErrorContext(msgCtx, "failed to process rabbitmq message", "queue", queueName, "error", err)
		_ = r.Nack(d.DeliveryTag, false, false) // send to DLQ / drop instead of infinite requeue loop
		return
	}

	if err := r.Ack(d.DeliveryTag, false); err != nil {
		slog.ErrorContext(msgCtx, "failed to ack rabbitmq message", "queue", queueName, "error", err)
	}
}

// Close gracefully tears down the network channels and sockets
func (r *RMQClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.isClosed = true
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (c AMQPHeaderCarrier) Get(key string) string {
	if val, ok := c[key]; ok {
		return fmt.Sprint(val)
	}
	return ""
}

func (c AMQPHeaderCarrier) Set(key, val string) {
	c[key] = val
}

func (c AMQPHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
