package order

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	Create(ctx context.Context, o Order) (Order, error)
	GetByID(ctx context.Context, id string) (Order, error)
	GetByIdemKey(ctx context.Context, key string) (Order, error)
}

type mongoRepo struct {
	col *mongo.Collection
}

func NewRepository(col *mongo.Collection) Repository {
	return &mongoRepo{col: col}
}

func (r *mongoRepo) Create(ctx context.Context, o Order) (Order, error) {
	o.CreatedAt = time.Now().UTC()
	_, err := r.col.InsertOne(ctx, o)
	return o, err
}

func (r *mongoRepo) GetByID(ctx context.Context, id string) (Order, error) {
	var out Order
	err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&out)
	return out, err
}

func (r *mongoRepo) GetByIdemKey(ctx context.Context, key string) (Order, error) {
	if key == "" {
		return Order{}, errors.New("empty idempotency key")
	}
	var out Order
	err := r.col.FindOne(ctx, bson.M{"idemKey": key}).Decode(&out)
	return out, err
}
