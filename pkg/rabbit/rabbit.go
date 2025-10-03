package rabbit

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Publisher struct {
	conn     *amqp091.Connection
	ch       *amqp091.Channel
	exchange string
	log      zerolog.Logger
}

func Connect(ctx context.Context, uri, exchange string, log zerolog.Logger) (*Publisher, error) {
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(
		exchange, "topic",
		true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	log.Info().Msg("connected to RabbitMQ & exchange declared")

	return &Publisher{conn: conn, ch: ch, exchange: exchange, log: log}, nil
}

func ConnectWithRetry(ctx context.Context, uri, exchange string, log zerolog.Logger, retries int, delay time.Duration) (*Publisher, error) {
    var lastErr error
    for i := 0; i < retries; i++ {
        pub, err := Connect(ctx, uri, exchange, log)
        if err == nil {
            return pub, nil
        }
        lastErr = err
        log.Warn().Err(err).Int("try", i+1).Msg("retrying RabbitMQ connection")
        time.Sleep(delay)
    }
    return nil, lastErr
}

func (p *Publisher) PublishJSON(ctx context.Context, routingKey string, body []byte) error {
	confirm := p.ch.NotifyPublish(make(chan amqp091.Confirmation, 1))

	if err := p.ch.PublishWithContext(ctx,
		p.exchange, routingKey, false, false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	); err != nil {
		return err
	}

	select {
	case c := <-confirm:
		if !c.Ack {
			return amqp091.ErrClosed
		}
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
	return nil
}

func (p *Publisher) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
