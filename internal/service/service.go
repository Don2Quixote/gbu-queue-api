package service

import (
	"context"
	"sync"

	"gbu-queue-api/internal/entity"

	"gbu-queue-api/pkg/logger"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

// Service is stuct that incapsulates business-logic's dependencies (interfaces).
type Service struct {
	posts Posts
	ctrls []Controller
	log   logger.Logger

	cast *broadcast
}

// New returns service with main business-logic.
// It can use multiple controllers.
func New(posts Posts, ctrls []Controller, log logger.Logger) *Service {
	return &Service{
		posts: posts,
		ctrls: ctrls,
		log:   log,

		cast: &broadcast{
			listeners: make(map[chan<- entity.Post]struct{}),
			mu:        &sync.Mutex{},
		},
	}
}

// Launch launches controllers and event's consuming.
// Blocks until context closed or launching error happened.
// Returns nil error if context closed.
func (s *Service) Launch(ctx context.Context) error {
	if len(s.ctrls) == 0 {
		return errors.New("0 controllers")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	posts, err := s.posts.Consume(ctx)
	if err != nil {
		return errors.Wrap(err, "consume posts")
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(s.ctrls) + 1) // Should wait all controllers + posts channel's close.

	handleNewPosts := func() {
		defer wg.Done()

		for post := range posts {
			s.log.Infof("new blog post %q - %s for %d listeners(s)", post.Title, post.URL, s.cast.listenersCount())
			s.cast.publish(post)
		}
	}
	go handleNewPosts()

	s.log.Infof("starting %d controller(s)", len(s.ctrls))

	// As controllers are launched in goroutines their error should be
	// registered atomically.
	ctrlErr := atomic.NewError(nil)

	for i, ctrl := range s.ctrls {
		go func(i int, ctrl Controller) {
			defer wg.Done()

			ctrl.RegisterConsume(s.Consume)
			err := ctrl.Launch(ctx)
			if err != nil {
				ctrlErr.Store(err)
				cancel()
			}

			s.log.Infof("controller (%d) stopped", i+1)
		}(i, ctrl)
	}

	wg.Wait()

	if ctrlErr.Load() != nil {
		return errors.Wrap(ctrlErr.Load(), "launch one of controllers")
	}

	return nil
}

// Consume returns channel with new posts that will be closed once context is done.
func (s *Service) Consume(ctx context.Context) (<-chan entity.Post, error) {
	posts := s.cast.subscribe()

	handleCtxClose := func() {
		<-ctx.Done()
		s.cast.unsubscribe(posts)
		close(posts)
	}
	go handleCtxClose()

	return posts, nil
}
