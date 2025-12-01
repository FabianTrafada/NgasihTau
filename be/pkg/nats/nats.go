// Package nats provides a NATS JetStream client wrapper for event publishing and subscribing.
// Implements requirement 10: API Gateway & Service Communication.
package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Config holds NATS connection configuration.
type Config struct {
	URL       string
	ClusterID string
	ClientID  string
}

// Client wraps NATS JetStream functionality.
type Client struct {
	conn      *nats.Conn
	js        jetstream.JetStream
	clientID  string
	clusterID string
}

// CloudEvent represents a CloudEvents formatted event.
// See: https://cloudevents.io/
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	Type            string          `json:"type"`
	Source          string          `json:"source"`
	ID              string          `json:"id"`
	Time            time.Time       `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

// NewClient creates a new NATS client with JetStream support.
func NewClient(cfg Config) (*Client, error) {
	// Connect to NATS
	conn, err := nats.Connect(cfg.URL,
		nats.Name(cfg.ClientID),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1), // Unlimited reconnects
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &Client{
		conn:      conn,
		js:        js,
		clientID:  cfg.ClientID,
		clusterID: cfg.ClusterID,
	}, nil
}

// Close closes the NATS connection.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// IsConnected returns true if the client is connected to NATS.
func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// Publish publishes an event to the specified subject using CloudEvents format.
func (c *Client) Publish(ctx context.Context, subject string, eventType string, data interface{}) error {
	// Marshal the data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Create CloudEvent
	event := CloudEvent{
		SpecVersion:     "1.0",
		Type:            eventType,
		Source:          c.clientID,
		ID:              uuid.New().String(),
		Time:            time.Now().UTC(),
		DataContentType: "application/json",
		Data:            dataBytes,
	}

	// Marshal the event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	// Publish to JetStream
	_, err = c.js.Publish(ctx, subject, eventBytes)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// PublishRaw publishes raw bytes to the specified subject.
func (c *Client) PublishRaw(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish raw event: %w", err)
	}
	return nil
}

// SubscribeConfig holds configuration for a subscription.
type SubscribeConfig struct {
	// Stream is the name of the JetStream stream to subscribe to.
	Stream string
	// Consumer is the durable consumer name.
	Consumer string
	// Subjects is the list of subjects to subscribe to.
	Subjects []string
	// MaxDeliver is the maximum number of delivery attempts.
	MaxDeliver int
	// AckWait is the time to wait for an acknowledgment.
	AckWait time.Duration
	// MaxAckPending is the maximum number of pending acknowledgments.
	MaxAckPending int
}

// DefaultSubscribeConfig returns a default subscription configuration.
func DefaultSubscribeConfig(stream, consumer string, subjects []string) SubscribeConfig {
	return SubscribeConfig{
		Stream:        stream,
		Consumer:      consumer,
		Subjects:      subjects,
		MaxDeliver:    3,
		AckWait:       30 * time.Second,
		MaxAckPending: 100,
	}
}

// MessageHandler is a function that handles incoming messages.
// It receives the CloudEvent and should return an error if processing fails.
// If nil is returned, the message will be acknowledged.
type MessageHandler func(ctx context.Context, event CloudEvent) error

// Subscription represents an active subscription to a NATS stream.
type Subscription struct {
	consumer jetstream.Consumer
	ctx      context.Context
	cancel   context.CancelFunc
}

// Stop stops the subscription.
func (s *Subscription) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// EnsureStream creates or updates a JetStream stream with the given configuration.
func (c *Client) EnsureStream(ctx context.Context, name string, subjects []string) error {
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        name,
		Description: fmt.Sprintf("Stream for %s events", name),
		Subjects:    subjects,
		Retention:   jetstream.WorkQueuePolicy,
		MaxAge:      7 * 24 * time.Hour, // 7 days retention
		Storage:     jetstream.FileStorage,
		Replicas:    1,
		Discard:     jetstream.DiscardOld,
	})
	if err != nil {
		return fmt.Errorf("failed to create/update stream %s: %w", name, err)
	}
	return nil
}

// Subscribe creates a durable subscription to the specified stream.
// The handler function is called for each message received.
// Returns a Subscription that can be used to stop the subscription.
func (c *Client) Subscribe(ctx context.Context, cfg SubscribeConfig, handler MessageHandler) (*Subscription, error) {
	// Get or create the consumer
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, cfg.Stream, jetstream.ConsumerConfig{
		Durable:        cfg.Consumer,
		AckPolicy:      jetstream.AckExplicitPolicy,
		AckWait:        cfg.AckWait,
		MaxDeliver:     cfg.MaxDeliver,
		MaxAckPending:  cfg.MaxAckPending,
		FilterSubjects: cfg.Subjects,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer %s: %w", cfg.Consumer, err)
	}

	// Create a cancellable context for the subscription
	subCtx, cancel := context.WithCancel(ctx)

	sub := &Subscription{
		consumer: consumer,
		ctx:      subCtx,
		cancel:   cancel,
	}

	// Start consuming messages in a goroutine
	go c.consumeMessages(subCtx, consumer, handler)

	return sub, nil
}

// consumeMessages continuously fetches and processes messages from the consumer.
func (c *Client) consumeMessages(ctx context.Context, consumer jetstream.Consumer, handler MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Fetch messages with a timeout
			msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
			if err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					return
				}
				// Log error and continue (transient errors are expected)
				continue
			}

			for msg := range msgs.Messages() {
				c.processMessage(ctx, msg, handler)
			}

			// Check for fetch errors
			if msgs.Error() != nil && ctx.Err() == nil {
				// Log error but continue
				continue
			}
		}
	}
}

// processMessage processes a single message and handles acknowledgment.
func (c *Client) processMessage(ctx context.Context, msg jetstream.Msg, handler MessageHandler) {
	// Parse the CloudEvent
	var event CloudEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		// If we can't parse the message, NAK it for redelivery
		// After max retries, it will go to dead letter
		_ = msg.Nak()
		return
	}

	// Call the handler
	if err := handler(ctx, event); err != nil {
		// Handler failed, NAK for redelivery
		_ = msg.Nak()
		return
	}

	// Success, acknowledge the message
	_ = msg.Ack()
}

// SubscribeSimple creates a simple push-based subscription using callbacks.
// This is a simpler alternative to Subscribe for cases where you don't need
// fine-grained control over message fetching.
func (c *Client) SubscribeSimple(ctx context.Context, cfg SubscribeConfig, handler MessageHandler) (*Subscription, error) {
	// Get or create the consumer
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, cfg.Stream, jetstream.ConsumerConfig{
		Durable:        cfg.Consumer,
		AckPolicy:      jetstream.AckExplicitPolicy,
		AckWait:        cfg.AckWait,
		MaxDeliver:     cfg.MaxDeliver,
		MaxAckPending:  cfg.MaxAckPending,
		FilterSubjects: cfg.Subjects,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer %s: %w", cfg.Consumer, err)
	}

	// Create a cancellable context for the subscription
	subCtx, cancel := context.WithCancel(ctx)

	// Start consuming with Consume (push-based)
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		c.processMessage(subCtx, msg, handler)
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	// Create subscription with cleanup
	sub := &Subscription{
		consumer: consumer,
		ctx:      subCtx,
		cancel: func() {
			cons.Stop()
			cancel()
		},
	}

	return sub, nil
}

// ParseEventData parses the Data field of a CloudEvent into the specified type.
func ParseEventData[T any](event CloudEvent) (T, error) {
	var data T
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return data, fmt.Errorf("failed to parse event data: %w", err)
	}
	return data, nil
}
