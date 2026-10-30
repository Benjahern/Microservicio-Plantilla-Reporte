package database

import (
	"context"
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

func Connect(ctx context.Context, c config.Config) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(c.MongoURI))
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, c.MongoTimeout)
	defer cancel()
	if err = client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return client, nil
}
func Timeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}
