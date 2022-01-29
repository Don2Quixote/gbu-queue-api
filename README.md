# Go Blog Updates - Queue API
This service consumes events about new posts in go blog ([go.dev](https://go.dev)) from message broker ([rabbitmq](https://www.rabbitmq.com/)) ([gbu-scanner service](https://github.com/don2quixote/gbu-scanner) publishes these events) and sends notifications to [websocket](https://datatracker.ietf.org/doc/html/rfc6455) and [grpc streams](https://grpc.io/) consumers.


## ENV Configuration:
| name                   | type   | description                                                                        |
| ---------------------- | ------ | ---------------------------------------------------------------------------------- |
| WEBSOCKET_PORT         | string | Port to launch http server with websocket handler on                               |
| GRPC_PORT              | string | Port to launch grpc server with streaming method on                                |
| RABBIT_HOST            | string | Rabbit host                                                                        |
| RABBIT_USER            | string | Rabbit user                                                                        |
| RABBIT_PASS            | string | Rabbit password                                                                    |
| RABBIT_VHOST           | string | Rabbit vhost                                                                       |
| RABBIT_AMQPS           | bool   | Flag to use amqps protocol instead of amqp                                         |
| RABBIT_RECONNECT_DELAY | int    | Delay (seconds) before attempting to reconnect to rabbit after loosing connection  |

Env template for sourcing is [deployments/local.env](deployments/local.env)
```
$ source deployments/local.env
```

## Makefile commands:
| name   | description                                                                            |
| ------ | -------------------------------------------------------------------------------------- |
| lint   | Runs linters                                                                           |
| test   | Runs tests, but there are no tests                                                     |
| docker | Builds docker image                                                                    |
| run    | Sources env variables from [deployments/local.env](deployments/local.env) and runs app |
| stat   | Prints stats information about project (packages, files, lines, chars count)           |

Direcotry [scripts](/scripts) contains scripts which invoked from [Makefile](Makefile)