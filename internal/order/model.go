package order

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Item struct {
	ProductID string `json:"productId" bson:"productId" validate:"required"`
	Quantity  int    `json:"quantity"  bson:"quantity"  validate:"gt=0"`
}

type Order struct {
	ID             string    `json:"id"        bson:"id"`
	UserID         string    `json:"userId"    bson:"userId" validate:"required"`
	Items          []Item    `json:"items"     bson:"items"  validate:"required,dive"`
	Status         string    `json:"status"    bson:"status"`
	IdempotencyKey string    `json:"-"         bson:"idemKey,omitempty"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
}

var validate = validator.New()

func (o *Order) Validate() error {
	return validate.Struct(o)
}

const (
	StatusPending = "pending"
	StatusPaid    = "paid"
	StatusCancel  = "cancelled"
)
