package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func (c *Controller) handleConsume(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to websocket.
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.log.Warn(errors.Wrap(err, "can't upgrade connection"))
		return
	}
	defer conn.Close()

	c.log.Info("new consumer")

	// posts chan will be closed once request's context will be done.
	posts, err := c.methods.consume(r.Context())
	if err != nil {
		c.log.Error(errors.Wrap(err, "can't get posts channel"))
		return
	}

	// Send each new post to the client.
	handleNewPosts := func() {
		for post := range posts {
			err := conn.WriteJSON(post)
			if err != nil {
				c.log.Error(errors.Wrap(err, "can't send post"))
			}
		}
	}
	go handleNewPosts()

	// If client doesn't send messages - it's probably dead.
	heartbeatsCount := 0
	heartbeats := make(chan struct{})
	handleHeartbeats := func() {
		isAlive := true
		for isAlive {
			select {
			case <-heartbeats:
				heartbeatsCount++ // It's not required to use atomic increment here.
				continue
			case <-time.After(time.Minute):
				warn := "closing connection due to no heartbeats"
				c.log.Warn(warn)

				err := conn.WriteMessage(websocket.TextMessage, []byte(warn))
				if err != nil {
					c.log.Warn(errors.Wrap(err, "can't write message to conn"))
				}

				conn.Close()

				isAlive = false
			case <-c.ctx.Done(): // r.Context() is not closed if server shutted down (hijacked connection) so another context.
				warn := "closing connection due to server's shutdown"
				c.log.Info(warn)

				err := conn.WriteMessage(websocket.TextMessage, []byte(warn))
				if err != nil {
					c.log.Error(errors.Wrap(err, "can't write message to conn"))
				}

				conn.Close()

				isAlive = false
			case <-r.Context().Done():
				isAlive = false
			}

		}
	}
	go handleHeartbeats()

	// Reading messages as heartbeats. If error reading message occurred -
	// connection considered as closed.
	start := time.Now()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			c.log.Infof("connection closed after %v with %d heartbeat(s)", time.Since(start), heartbeatsCount)
			break
		}

		heartbeats <- struct{}{}
	}
}
