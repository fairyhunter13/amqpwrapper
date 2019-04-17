package rabbitmq

import (
	"time"
)

const (
	//Producer defines the type for Producer client of rabbitmq.
	Producer uint64 = iota + 1
	//Consumer defines the type for Consumer client of rabbitmq.
	Consumer
)

const (
	//DefaultHeartbeat defines the default heartbeat used in the config to dial rabbitmq.
	DefaultHeartbeat = 10 * time.Second
	//DefaultLocale set the default locale for the connection to rabbitmq.
	DefaultLocale = "en_US"
)

const (
	//Open defines the opened state for the connection.
	Open uint32 = iota
	//Closed defines the closed state for the connection.
	Closed
)
