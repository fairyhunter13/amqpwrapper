/*
	Package amqpwrapper manage the connection and channel from streadway/amqp.

	Amqpwrapper wraps the connection and channel from streadway/amqp.
	This package manages auto reconnection and channel lifecycle.
	This package ensures two connection will be initialized for each type (producers and consumers).
	This has to be done because initialization of tcp connection is very costly and slowing performance greatly in production.

*/
package amqpwrapper
