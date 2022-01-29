package app

// appConfig is struct for parsing ENV configuration.
type appConfig struct {
	// WebsocketPort is port on which http server with websocket handler launched.
	WebsocketPort int `config:"WEBSOCKET_PORT,required"`
	// GRPCPort is port on which grpc server with stream handler launched.
	GRPCPort int `config:"GRPC_PORT,required"`
	// RabbitHost is host of rabbitmq.
	RabbitHost string `config:"RABBIT_HOST,required"`
	// RabbitUser is user for rabbitmq.
	RabbitUser string `config:"RABBIT_USER"`
	// RabbitPass is password for rabbitmq.
	RabbitPass string `config:"RABBIT_PASS"`
	// RabbitVhost is vhost in rabbitmq to connect.
	RabbitVhost string `config:"RABBIT_VHOST"`
	// RabbitAmqps flag shows should amqps protocol be used instead of amqp or not.
	RabbitAmqps bool `config:"RABBIT_AMQPS"`
	// RabbitReconnectDelay is delay (in seconds) before attempting to reconnect to rabbit after loosing connection.
	RabbitReconnectDelay int `config:"RABBIT_RECONNECT_DELAY,required"`
}
