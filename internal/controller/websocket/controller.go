package websocket

import (
	"context"
	"fmt"
	"net/http"

	"gbu-queue-api/internal/service"

	"gbu-queue-api/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Controller is implementation for service.Controller interface via HTTP server.
type Controller struct {
	port int
	log  logger.Logger

	methods struct {
		consume consumeFunc
	}
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	ctx      context.Context
}

var _ service.Controller = &Controller{}

// New returns service.Controller implementation.
func New(port int, log logger.Logger) *Controller {
	return &Controller{
		port: port,
		log:  logger.Extend(log, "controller", "websocket"),

		mux: http.NewServeMux(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  bufSize,
			WriteBufferSize: bufSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
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

	c.handleFunc("/consume", c.handleConsume)

	// Handler for everything that didn't match /consume path to return NotFound status code.
	c.handleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.log.Warnf("request to not registered path %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("websocket connection available at /consume path"))
		if err != nil {
			c.log.Error(errors.Wrap(err, "can't write response"))
		}
	})

	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", c.port),
		Handler:     c.mux,
		ReadTimeout: readTimeout,
	}

	shutdown := make(chan error)

	// Should gracefully shutdown server once context will be closed.
	handleCtxClose := func() {
		<-ctx.Done()

		c.log.Info("shutting down server")

		cancel()

		err := server.Shutdown(context.Background())
		if err != nil {
			c.log.Error(errors.Wrap(err, "can't shutdown server"))
		}

		shutdown <- nil
	}
	go handleCtxClose()

	startServer := func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdown <- errors.Wrap(err, "can't start server")
		}
	}
	go startServer()

	// nil if context closed, not nil otherwise.
	return <-shutdown
}

func (c *Controller) RegisterConsume(consume consumeFunc) {
	c.methods.consume = consume
}

// handleFunc registers http handler for given pattern.
// If given handler is nil then handler that writes http.StatusNotImplemented code registers.
func (c *Controller) handleFunc(pattern string, handler http.HandlerFunc) {
	if handler == nil {
		c.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
		})
		return
	}

	c.mux.HandleFunc(pattern, handler)
}
