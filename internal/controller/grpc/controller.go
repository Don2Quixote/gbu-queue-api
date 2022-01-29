package grpc

import (
	"context"
	"fmt"
	"net"

	"gbu-queue-api/internal/controller/grpc/proto"
	"gbu-queue-api/internal/service"

	"gbu-queue-api/pkg/logger"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Controller is implementation for service.Controller interface via HTTP server.
type Controller struct {
	port int
	log  logger.Logger

	methods struct {
		consume consumeFunc
	}

	// ctx is context which used to determine when server
	// is shutting down and when all consumers should be disconnected.
	// Initialized in Launch method.
	ctx context.Context

	proto.UnimplementedPostsServer
}

var _ service.Controller = &Controller{}
var _ proto.PostsServer = &Controller{}

// New returns service.Controller implementation.
func New(port int, log logger.Logger) *Controller {
	return &Controller{
		port: port,
		log:  logger.Extend(log, "controller", "grpc"),

		ctx: nil, // Initialized in Launch method.
	}
}

func (c *Controller) Launch(ctx context.Context) error {
	c.log.Info("launching server")

	// In Consume method streams won't be disconnected on
	// server's shutdown automatically. So added this c.ctx to
	// control their disconnection on context's close.
	// context.Background() used instead of ctx to guarantee
	// that firstly "shutting down server" will be logged, only then
	// cancel() will be called and clients will be disconnected.
	var cancel func()
	c.ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", c.port))
	if err != nil {
		return errors.Wrapf(err, "can't listen port %d", c.port)
	}
	defer lis.Close()

	server := grpc.NewServer()

	proto.RegisterPostsServer(server, c)

	shutdown := make(chan error)

	// Should gracefully shutdown server once context will be closed.
	handleCtxClose := func() {
		<-ctx.Done()
		c.log.Info("shutting down server")
		cancel()              // Cancel context client's attached to.
		server.GracefulStop() // Wait clients to receive cancel()'s signal and disconnect.
		shutdown <- nil
	}
	go handleCtxClose()

	serve := func() {
		err = server.Serve(lis)
		if err != nil {
			shutdown <- errors.Wrap(err, "can't serve")
		}
	}
	go serve()

	// nil if context closed, not nil otherwise.
	return <-shutdown
}

func (c *Controller) RegisterConsume(consume consumeFunc) {
	c.methods.consume = consume
}
