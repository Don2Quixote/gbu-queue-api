package service

import (
	"context"

	"gbu-queue-api/internal/entity"
)

// Posts is interface for interacting with message broker
// from where events about new posts in blog come.
type Posts interface {
	// Consume returns channel to which new blog's posts will be sent.
	// Returned chan should be closed when context will be closed.
	Consume(ctx context.Context) (<-chan entity.Post, error)
}

// Controller is interface to register controllers for the service.
type Controller interface {
	// Launch launches controller.
	// It should return nil error if context closed.
	Launch(ctx context.Context) error
	// RegisterConsume registers handler for service's consume method.
	RegisterConsume(consumeFunc)
}
