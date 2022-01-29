package grpc

import (
	"time"

	"gbu-queue-api/internal/controller/grpc/proto"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Controller) Consume(e *proto.Empty, stream proto.Posts_ConsumeServer) error {
	c.log.Info("new consumer")

	start := time.Now()

	// posts chan will be closed once request's context will be done.
	posts, err := c.methods.consume(stream.Context())
	if err != nil {
		return errors.Wrap(err, "can't consume")
	}

	// Send each new post to the client.
	handleNewPosts := func() {
		for post := range posts {
			err := stream.Send(&proto.Post{
				Title:   post.Title,
				Date:    timestamppb.New(post.Date),
				Author:  post.Author,
				Summary: post.Summary,
				Url:     post.URL,
			})
			if err != nil {
				c.log.Error(errors.Wrap(err, "can't send post"))
			}
		}
	}
	go handleNewPosts()

	// Wait eiterh client's disconnection or server's shutdown.
	select {
	case <-stream.Context().Done():
		c.log.Infof("consumer disconnected after %v", time.Since(start))
	case <-c.ctx.Done():
		c.log.Infof("consumer disconnected after %v due to server's shutdown", time.Since(start))
	}

	return nil
}
