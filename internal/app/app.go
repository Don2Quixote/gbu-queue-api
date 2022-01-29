package app

import (
	"context"
	"time"

	"gbu-queue-api/internal/controller/grpc"
	"gbu-queue-api/internal/controller/websocket"
	"gbu-queue-api/internal/posts"
	"gbu-queue-api/internal/service"

	"gbu-queue-api/pkg/config"
	"gbu-queue-api/pkg/logger"

	"github.com/pkg/errors"
)

func Run(ctx context.Context, log logger.Logger) error {
	log.Info("starting app")

	var cfg appConfig
	err := config.Parse(&cfg)
	if err != nil {
		return errors.Wrap(err, "can't parse config")
	}

	posts := posts.New(posts.RabbitConfig{
		Host:           cfg.RabbitHost,
		User:           cfg.RabbitUser,
		Pass:           cfg.RabbitPass,
		Vhost:          cfg.RabbitVhost,
		Amqps:          cfg.RabbitAmqps,
		ReconnectDelay: time.Duration(cfg.RabbitReconnectDelay) * time.Second,
	}, log)

	err = posts.Init(ctx, ctx)
	if err != nil {
		return errors.Wrap(err, "can't init posts")
	}

	wsCtrl := websocket.New(cfg.WebsocketPort, log)
	grpcCtrl := grpc.New(cfg.GRPCPort, log)

	service := service.New(posts, []service.Controller{wsCtrl, grpcCtrl}, log)

	err = service.Launch(ctx)
	if err != nil {
		return errors.Wrap(err, "can't launch service")
	}

	log.Info("app finished")

	return nil
}
