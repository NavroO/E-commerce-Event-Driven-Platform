package mongo

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri string, log zerolog.Logger) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := cl.Ping(ctx, nil); err != nil {
		return nil, err
	}
	log.Info().Msg("connected to MongoDB")
	return cl, nil
}

func EnsureIndexes(ctx context.Context, db *mongo.Database, log zerolog.Logger) error {
	orders := db.Collection("orders")

	_, err := orders.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_order_id"),
	})
	if err != nil {
		return err
	}

	_, err = orders.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "idemKey", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_idem_key"),
	})
	if err != nil {
		return err
	}

	log.Info().Msg("mongo indexes ensured")
	return nil
}
