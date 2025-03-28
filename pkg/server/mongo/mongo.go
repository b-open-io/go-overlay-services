package mongo

import (
	"context"
	"time"

	"github.com/4chain-ag/go-overlay-services/pkg/server/config"
	"github.com/gookit/slog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Database *mongo.Database
	client   *mongo.Client
}

func New(cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI)
	mongoClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(cfg.Mongo.Database)
	slog.Info("MongoDB connected to:", cfg.Mongo.URI, "using database:", cfg.Mongo.Database)

	return &Client{
		Database: db,
		client:   mongoClient,
	}, nil
}
