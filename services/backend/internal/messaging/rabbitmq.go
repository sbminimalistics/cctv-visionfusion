package messaging

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RMQClient handles your RabbitMQ connection and channel pool lifecycles safely
type RMQClient struct {
	url      string
	conn     *amqp.Connection
	ch       *amqp.Channel
	mu       sync.RWMutex
	isClosed bool
}

// NewRMQClient initializes the client structure configuration
func NewRMQClient(user, password, host string, port int, vhost string) *RMQClient {
	cleanVhost := strings.TrimPrefix(vhost, "/")

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, password, host, port, cleanVhost)
	fmt.Println("RMQClient url:", url)
	return &RMQClient{url: url}
}

// Connect establishes the network socket and handles recovery loops
func (r *RMQClient) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Println("RMQClient dialing url:", r.url)
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

		// Backoff reconnection loop until the broker is recovered
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

	if r.ch == nil || r.conn.IsClosed() {
		return fmt.Errorf("cannot publish message: rabbitmq client is disconnected")
	}

	// Dispatch message targeting standard AMQP 0-9-1 configurations
	err := r.ch.PublishWithContext(ctx,
		exchangeName, // Exchange name string
		routingKey,   // Target queue routing identifier
		false,        // Mandatory flag
		false,        // Immediate flag
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Forces message persistence across broker restarts
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("message publication failed: %w", err)
	}

	return nil
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
