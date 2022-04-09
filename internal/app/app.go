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

// Run runs app. If returned error is not nil, program exited
// unexpectedly and non-zero code should be returned (os.Exit(1) or log.Fatal(...)).
func Run(ctx context.Context, log logger.Logger) error {
	log.Info("starting app")

	var cfg appConfig
	err := config.Parse(&cfg)
	if err != nil {
		return errors.Wrap(err, "parse config")
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
		return errors.Wrap(err, "init posts")
	}

	wsCtrl := websocket.New(cfg.WebsocketPort, log)
	grpcCtrl := grpc.New(cfg.GRPCPort, log)

	service := service.New(posts, []service.Controller{wsCtrl, grpcCtrl}, log)

	err = service.Launch(ctx)
	if err != nil {
		return errors.Wrap(err, "launch service")
	}

	log.Info("app finished")

	return nil
}
