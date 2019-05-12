package amqpwrapper

import "github.com/streadway/amqp"

type (

	//InitializeChannel is a closure to init all actions to channel like declaring exchange
	//, binding to queue, declaring queue, and etc.
	InitializeChannel func(*amqp.Channel) error

	//IChannelManager defines the custom channel contract.
	IChannelManager interface {
		Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
		Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
		IsClosed() (result bool)
	}

	//Channel defines the custom channel for rabbitmq.
	Channel struct {
		fn        InitializeChannel
		innerChan *amqp.Channel
	}

	//ChannelManager defines the manager for channel to ease the publish and consume for a message.
	ChannelManager struct {
		connMgr  IConnectionManager
		key      string
		typeChan uint64
	}
)

//NewChannelManager creates new channel manager for the given key and channel type.
func NewChannelManager(key string, typeChan uint64, fn InitializeChannel, connMgr IConnectionManager) (mgr IChannelManager, err error) {
	if connMgr == nil {
		err = ErrNilArg
		return
	}
	if isNotValidTypeChan(typeChan) || isNotValidKey(key) {
		err = ErrInvalidArgs
		return
	}
	amqpChan, err := connMgr.CreateChannel(typeChan)
	if err != nil {
		return
	}
	args := InitArgs{
		Key:      key,
		TypeChan: typeChan,
		Channel:  amqpChan,
	}
	err = connMgr.InitChannel(fn, args)
	if err != nil {
		return
	}
	chanMgr := &ChannelManager{
		connMgr:  connMgr,
		key:      key,
		typeChan: typeChan,
	}
	mgr = chanMgr
	return
}

//Publish defines the Publish function inside amqp.Channel.
//See github.com/streadway/amqp for more details.
func (m *ChannelManager) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) (err error) {
	mqChan, err := m.connMgr.GetChannel(m.key, m.typeChan)
	if err != nil {
		return
	}
	err = mqChan.Publish(exchange, key, mandatory, immediate, msg)
	return
}

//Consume defines the Consume function inside amqp.Channel.
//See github.com/streadway/amqp for more details.
func (m *ChannelManager) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (deliveryChan <-chan amqp.Delivery, err error) {
	mqChan, err := m.connMgr.GetChannel(m.key, m.typeChan)
	if err != nil {
		return
	}
	deliveryChan, err = mqChan.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return
}

//IsClosed check the status of connection status of connection manager.
func (m *ChannelManager) IsClosed() bool {
	return m.connMgr.IsClosed()
}
