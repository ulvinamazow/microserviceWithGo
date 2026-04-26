package database

import (
	"context"
	"fmt"
	"time"

	"github.org/ulvinamazow/microservice_with_go/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoClient struct {
	client *mongo.Client
	DB     *mongo.Database
}

func Connect(ctx context.Context, cfg config.MongoDBConfig) (*MongoClient, error) {
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTimeSec) * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("Error while connection database: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("Ping time error occurred: %w", err)
	}

	db := client.Database(cfg.Database)

	zap.L().Info("MongoDB connection established successfully",
		zap.String("database", cfg.Database),
		zap.Uint64("maxPoolSize", cfg.MinPoolSize),
	)

	return &MongoClient{
		client: client,
		DB:     db,
	}, nil
}

func (mc *MongoClient) Close(ctx context.Context) error {
	if mc.client == nil {
		return nil
	}

	if err := mc.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("Mongo Disconnect: %w", err)
	}

	zap.L().Info("MongoDB connection terminated")
	return nil
}
