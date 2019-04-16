package rabbitmq

import "errors"

var (
	//ErrChannelNotFound defines error if the custom channel is not found in the map.
	ErrChannelNotFound = errors.New("Channel is not found in the channel map")
	//ErrNilArg defines the error if an argument is nil.
	ErrNilArg = errors.New("Argument must not be nil")
	//ErrInvalidArgs defines the error if arguments are not valid.
	ErrInvalidArgs = errors.New("Arguments are not valid")
)
