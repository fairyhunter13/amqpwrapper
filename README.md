# Amqpwrapper
[![CircleCI](https://circleci.com/gh/fairyhunter13/amqpwrapper/tree/master.svg?style=svg)](https://circleci.com/gh/fairyhunter13/amqpwrapper/tree/master)
[![Coverage Status](https://coveralls.io/repos/github/fairyhunter13/amqpwrapper/badge.svg?branch=master)](https://coveralls.io/github/fairyhunter13/amqpwrapper?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/fairyhunter13/amqpwrapper)](https://goreportcard.com/report/github.com/fairyhunter13/amqpwrapper)

Amqwrapper is a library to wrap streadway/amqp. 
This library manages channel initialization and reconnection automatically. 
Since streadway/amqp doesn't provide the mechanism for auto reconnection, this library does this job and ensures that the topology still remains the same when the channel firstly initialized.
For more details, see the documentation of the [Api Reference](https://godoc.org/github.com/fairyhunter13/amqpwrapper) in here.

