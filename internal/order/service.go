package order

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Publisher interface {
	PublishJSON(ctx context.Context, routingKey string, body []byte) error
}

type Service struct {
	repo      Repository
	pub       Publisher
	logger    zerolog.Logger
	eventKey  string
	timeout   time.Duration
}

func NewService(r Repository, p Publisher, log zerolog.Logger) *Service {
	return &Service{
		repo:     r,
		pub:      p,
		logger:   log,
		eventKey: "order.created.v1",
		timeout:  5 * time.Second,
	}
}

type CreateOrderRequest struct {
	UserID         string `json:"userId"`
	Items          []Item `json:"items"`
	IdempotencyKey string `json:"-"`
}

type OrderCreatedEvent struct {
	OrderID    string    `json:"orderId"`
	UserID     string    `json:"userId"`
	Items      []Item    `json:"items"`
	OccurredAt time.Time `json:"occurredAt"`
	Version    int       `json:"version"`
}

func (s *Service) Create(ctx context.Context, req CreateOrderRequest) (Order, error) {
	if req.IdempotencyKey != "" {
		if existing, err := s.repo.GetByIdemKey(ctx, req.IdempotencyKey); err == nil && existing.ID != "" {
			return existing, nil
		}
	}

	o := Order{
		ID:             uuid.NewString(),
		UserID:         req.UserID,
		Items:          req.Items,
		Status:         StatusPending,
		IdempotencyKey: req.IdempotencyKey,
	}
	if err := o.Validate(); err != nil {
		return Order{}, err
	}

	cctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	created, err := s.repo.Create(cctx, o)
	if err != nil {
		return Order{}, err
	}

	evt := OrderCreatedEvent{
		OrderID:    created.ID,
		UserID:     created.UserID,
		Items:      created.Items,
		OccurredAt: time.Now().UTC(),
		Version:    1,
	}
	payload, _ := json.Marshal(evt)

	if err := s.pub.PublishJSON(cctx, s.eventKey, payload); err != nil {
		s.logger.Error().Err(err).Msg("failed to publish order.created")
	}

	return created, nil
}

func (s *Service) Get(ctx context.Context, id string) (Order, error) {
	if id == "" {
		return Order{}, errors.New("empty id")
	}
	return s.repo.GetByID(ctx, id)
}
