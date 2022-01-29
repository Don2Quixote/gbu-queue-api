package websocket

import (
	"context"

	"gbu-queue-api/internal/entity"
)

type consumeFunc = func(ctx context.Context) (<-chan entity.Post, error)
