package amqpwrapper

//go:generate mockery -all -inpkg -recursive -testonly

import (
	"sync"
	"sync/atomic"

	"github.com/streadway/amqp"
)

type (

	// IConnectionManager defines the contract to manage connection in this package.
	IConnectionManager interface {
		GetChannel(key string, typeChan uint64) (channel *amqp.Channel, err error)
		CreateChannel(typeChan uint64) (channel *amqp.Channel, err error)
		InitChannel(fn InitializeChannel, args InitArgs) (err error)
		InitChannelAndGet(fn InitializeChannel, args InitArgs) (channel *amqp.Channel, err error)
		Close() (err error)
		IsClosed() (result bool)
	}

	// ConnectionManager defines the manager for connection used in producer and consumer of RabbitMQ.
	ConnectionManager struct {
		isClosed uint32 // 1 means closed, 0 means not closed.
		wg       sync.WaitGroup

		// stores config and url for reconnect
		url    string
		config amqp.Config
		mutex  sync.RWMutex
		// mutex protects the following field.
		producer map[string]*Channel
		consumer map[string]*Channel
		prodConn *amqp.Connection
		consConn *amqp.Connection
		prodErr  chan *amqp.Error
		consErr  chan *amqp.Error
	}

	// InitArgs defines the arguments for InitChannel function in this package.
	InitArgs struct {
		Key      string
		TypeChan uint64
		Channel  *amqp.Channel
	}
)

// NewManager creates connection manager to be used to manage the lifecycle of connections.
func NewManager(url string, config amqp.Config) (manager IConnectionManager, err error) {
	if config.Heartbeat <= 0 {
		config.Heartbeat = DefaultHeartbeat
	}
	if config.Locale == "" {
		config.Locale = DefaultLocale
	}
	if url == "" {
		err = ErrInvalidArgs
		return
	}

	mgr := &ConnectionManager{
		url:      url,
		config:   config,
		producer: make(map[string]*Channel, 0),
		consumer: make(map[string]*Channel, 0),
	}

	err = mgr.connect()
	if err != nil {
		return
	}

	go mgr.reconnect()

	manager = mgr
	return
}

// CreateChannel creates the channel with connection that has been initialized before.
// CreateChannel initialize channel based on type of channel in the input.
func (p *ConnectionManager) CreateChannel(typeChan uint64) (channel *amqp.Channel, err error) {
	switch typeChan {
	case Producer:
		channel, err = p.prodConn.Channel()
	case Consumer:
		channel, err = p.consConn.Channel()
	default:
		err = ErrInvalidArgs
	}
	return
}

// GetChannel returns the channel that is stored inside the map in the manager.
func (p *ConnectionManager) GetChannel(key string, typeChan uint64) (channel *amqp.Channel, err error) {
	var (
		customChannel *Channel
		ok            bool
	)
	p.mutex.RLock()
	if typeChan == Producer {
		customChannel, ok = p.producer[key]
	} else if typeChan == Consumer {
		customChannel, ok = p.consumer[key]
	}
	p.mutex.RUnlock()
	if !ok {
		err = ErrChannelNotFound
		return
	}
	channel = customChannel.innerChan
	return
}

// InitChannel initialize channel with fn and add it to the map to recover or reinit.
func (p *ConnectionManager) InitChannel(fn InitializeChannel, args InitArgs) (err error) {
	if args.Channel == nil || fn == nil {
		err = ErrNilArg
		return
	}
	if p.isNotValidChannelArgs(args) {
		err = ErrInvalidArgs
		return
	}
	err = p.lockInitChannel(fn, args)
	return
}

// InitChannelAndGet initialize channel with fn and add it to the map to recover or reinit, then return the created channel.
func (p *ConnectionManager) InitChannelAndGet(fn InitializeChannel, args InitArgs) (ch *amqp.Channel, err error) {
	if fn == nil {
		err = ErrNilArg
		return
	}
	if p.isNotValidChannelArgs(args) {
		err = ErrInvalidArgs
		return
	}
	ch, err = p.initChannelAndGet(fn, args, true)
	return
}

func (p *ConnectionManager) initChannelAndGet(fn InitializeChannel, args InitArgs, withLock bool) (ch *amqp.Channel, err error) {
	if args.Channel == nil {
		args.Channel, err = p.CreateChannel(args.TypeChan)
		if err != nil {
			return
		}
	}
	if withLock {
		err = p.lockInitChannel(fn, args)
	} else {
		err = p.initChannel(fn, args)
	}
	if err != nil {
		return
	}
	ch = args.Channel
	return
}

func (p *ConnectionManager) lockInitChannel(fn InitializeChannel, args InitArgs) (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err = p.initChannel(fn, args)
	return err
}

// Close close the connection and channel.
func (p *ConnectionManager) Close() (err error) {
	atomic.StoreUint32(&p.isClosed, Closed)
	err = p.prodConn.Close()
	err = p.consConn.Close()
	return
}

// IsClosed defines the state of connection in manager.
func (p *ConnectionManager) IsClosed() bool {
	return atomic.LoadUint32(&p.isClosed) == Closed
}

// connect creates new connection and add close notifier to trigger reinit or reconnect.
func (p *ConnectionManager) connect() (err error) {
	err = p.connectProducer()
	if err != nil {
		return
	}
	err = p.connectConsumer()
	return
}

func (p *ConnectionManager) connectProducer() (err error) {
	conn, err := amqp.DialConfig(p.url, p.config)
	if err != nil {
		return
	}
	p.mutex.Lock()
	p.prodConn = conn
	p.prodErr = conn.NotifyClose(make(chan *amqp.Error, 1))
	p.mutex.Unlock()
	return
}

func (p *ConnectionManager) connectConsumer() (err error) {
	conn, err := amqp.DialConfig(p.url, p.config)
	if err != nil {
		return
	}
	p.mutex.Lock()
	p.consConn = conn
	p.consErr = conn.NotifyClose(make(chan *amqp.Error, 1))
	p.mutex.Unlock()
	return
}

func (p *ConnectionManager) initChannel(fn InitializeChannel, args InitArgs) (err error) {
	err = fn(args.Channel)
	if err != nil {
		return
	}
	customChannel := &Channel{
		fn:        fn,
		innerChan: args.Channel,
	}
	if args.TypeChan == Consumer {
		p.consumer[args.Key] = customChannel
	} else if args.TypeChan == Producer {
		p.producer[args.Key] = customChannel
	}
	return
}

func (p *ConnectionManager) isNotValidChannelArgs(args InitArgs) bool {
	return isNotValidKey(args.Key) || isNotValidTypeChan(args.TypeChan)
}

// reconnect does the reconnection of the rabbitmq manager.
func (p *ConnectionManager) reconnect() {
	p.wg.Add(1)
	// Reconnecting for producers.
	go func() {
		defer p.wg.Done()
		for {
			<-p.prodErr

			if atomic.LoadUint32(&p.isClosed) == Open {
				p.connectProducer()
				p.reinitProducer()
			} else {
				return
			}
		}
	}()
	p.wg.Add(1)
	// Reconnecting for consumers.
	go func() {
		defer p.wg.Done()
		for {
			<-p.consErr

			if atomic.LoadUint32(&p.isClosed) == Open {
				p.connectConsumer()
				p.reinitConsumer()
			} else {
				return
			}
		}
	}()
	p.wg.Wait()
	return
}

func (p *ConnectionManager) reinitConsumer() (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for key, customChannel := range p.consumer {
		customChannel.innerChan.Close()
		argsChan := InitArgs{
			Key:      key,
			TypeChan: Consumer,
		}
		_, err = p.initChannelAndGet(customChannel.fn, argsChan, false)
		if err != nil {
			return
		}
	}
	return
}

func (p *ConnectionManager) reinitProducer() (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for key, customChannel := range p.producer {
		customChannel.innerChan.Close()
		argsChan := InitArgs{
			Key:      key,
			TypeChan: Producer,
		}
		_, err = p.initChannelAndGet(customChannel.fn, argsChan, false)
		if err != nil {
			return
		}
	}
	return
}
