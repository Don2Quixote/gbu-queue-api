package posts

import (
	"context"
	"encoding/json"
	"sync"

	"gbu-queue-api/internal/entity"
	"gbu-queue-api/internal/service"

	"gbu-queue-api/pkg/logger"
	"gbu-queue-api/pkg/sleep"
	"gbu-queue-api/pkg/wrappers/rabbit"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Posts is implementation for service.Posts interface.
type Posts struct {
	rabbitConfig RabbitConfig
	log          logger.Logger

	rabbit    *amqp.Channel
	queueName string
	mu        *sync.Mutex
}

var _ service.Posts = &Posts{}

// New returns service.Posts implementation.
func New(rabbitConfig RabbitConfig, log logger.Logger) *Posts {
	return &Posts{
		rabbitConfig: rabbitConfig,
		log:          log,

		rabbit:    nil, // Initialized in Init method.
		queueName: "",  // Initialized in Init method.
		mu:        &sync.Mutex{},
	}
}

// Init connects to rabbit and gets rabbit channel, after what
// initializes rabbit's entiies like exchanges, queues etc.
// It also registers a handler for channel closed event to reconnect.
// Close handler uses processCtx for it's calls because ctx for Init's call
// can be another: for example, limited as WithTimeout.
func (p *Posts) Init(ctx, processCtx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cfg := p.rabbitConfig

	conn, err := rabbit.Dial(cfg.Host, cfg.User, cfg.Pass, cfg.Vhost, cfg.Amqps)
	if err != nil {
		return errors.Wrap(err, "can't connect to rabbit")
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "can't get rabbit channel")
	}

	err = ch.ExchangeDeclare(postsExchange, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "can't declare posts exchange")
	}

	queue, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return errors.Wrap(err, "can't declare exclusive queue")
	}

	p.log.Infof("declared rabbit queue %q", queue.Name)

	p.queueName = queue.Name

	// Exchange is fanout so no binding key required.
	err = ch.QueueBind(p.queueName, "", postsExchange, false, nil)
	if err != nil {
		return errors.Wrap(err, "can't bind exclusive queue to posts exchange")
	}

	errs := make(chan *amqp.Error)
	ch.NotifyClose(errs)

	handleChannelClose := func() {
		closeErr := <-errs // This chan will get a value when rabbit channel will be closed.

		p.log.Error(errors.Wrap(closeErr, "rabbit channel closed"))

		if !conn.IsClosed() {
			err := conn.Close()
			if err != nil {
				p.log.Error(errors.Wrap(err, "can't close rabbit connection"))
			}
		}

		for attempt, isConnected := 1, false; !isConnected; attempt++ {
			isCtxClosed := sleep.WithContext(processCtx, cfg.ReconnectDelay)
			if isCtxClosed {
				p.log.Info("could not re-init consumer until context closed")
				return
			}

			err := p.Init(processCtx, processCtx)
			if err != nil {
				p.log.Warn(errors.Wrapf(err, "can't re-init consumer (attempt #%d)", attempt))
				continue
			}

			isConnected = true
		}

		p.log.Info("reconnected to rabbit")
	}
	go handleChannelClose()

	p.rabbit = ch

	return nil
}

func (p *Posts) Consume(ctx context.Context) (<-chan entity.Post, error) {
	posts := make(chan entity.Post)

	messages, err := p.rabbit.Consume(p.queueName, "", true, true, false, false, nil)

	// waitReconnection tryies to consume from queue.
	// returns false if context closed before reconnected, true otherwise.
	waitReconnection := func() bool {
		// Loop until connection reestablished or context closed.
		for {
			isCtxClosed := sleep.WithContext(ctx, p.rabbitConfig.ReconnectDelay)
			if isCtxClosed {
				return false
			}

			// TODO: Guess it can be a data race with c.rabbit.
			messages, err = p.rabbit.Consume(p.queueName, "", true, true, false, false, nil)
			if err == nil {
				return true
			}
		}
	}

	handleMessage := func(message amqp.Delivery) error {
		var post entity.Post
		err := json.Unmarshal(message.Body, &post)
		if err != nil {
			return errors.Wrapf(err, "can't decode message %q", message.Body)
		}

		posts <- post

		return nil
	}

	handleMessages := func() {
		for {
			select {
			case message, ok := <-messages:
				// ok is false if messages chan is closed and reconnection needed.
				if !ok {
					isReconnected := waitReconnection()
					if !isReconnected {
						close(posts)
						return
					}
					continue
				}

				err := handleMessage(message)
				if err != nil {
					p.log.Error(errors.Wrap(err, "can't handle message"))
				}
			case <-ctx.Done():
				close(posts)
				return
			}
		}
	}
	go handleMessages()

	return posts, nil
}
