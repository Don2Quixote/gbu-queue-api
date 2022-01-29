package websocket

import "time"

// Who will configure these values 😂? I don't give a fuck - constants.

// bufSize is const for websocket.Upgrader's configuration.
const bufSize = 1024

// readTimeout is const for http server's read timeout.
const readTimeout = time.Second * 30
