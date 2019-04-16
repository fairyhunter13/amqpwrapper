package rabbitmq

import (
	"sync"
	"sync/atomic"

	"github.com/streadway/amqp"
)

type (

	//IConnectionManager defines the contract for the ConnectionManager.
	IConnectionManager interface {
		GetChannel(key string, typeChan uint64) (channel *amqp.Channel, err error)
		CreateChannel(typeChan uint64) (channel *amqp.Channel, err error)
		InitChannel(fn InitializeChannel, args InitArgs) (err error)
		Close() (err error)
		IsClosed() (result bool)
	}

	//ConnectionManager defines the manager for connection used in producer and consumer of RabbitMQ.
	ConnectionManager struct {
		isClosed uint32 //1 means closed, 0 means not closed.

		//stores config and url for reconnect
		url    string
		config amqp.Config
		mutex  *sync.RWMutex
		//mutex protects the following field.
		producer map[string]*Channel
		consumer map[string]*Channel
		prodConn *amqp.Connection
		consConn *amqp.Connection
		prodErr  chan *amqp.Error
		consErr  chan *amqp.Error
	}

	//InitArgs defines the arguments for InitChannel function in this package.
	InitArgs struct {
		Key      string
		TypeChan uint64
		Channel  *amqp.Channel
	}
)

//NewManager defines the manager for producer.
//This NewManager need to be tested with integration test.
func NewManager(url string, config amqp.Config) (manager IConnectionManager, err error) {
	if config.Heartbeat <= 0 {
		config.Heartbeat = defaultHeartbeat
	}
	if config.Locale == "" {
		config.Locale = defaultLocale
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
		mutex:    new(sync.RWMutex),
	}

	err = mgr.connect()
	if err != nil {
		return
	}

	go mgr.reconnect()

	manager = mgr
	return
}

//CreateChannel creates the channel with connection from inside the map.
func (p *ConnectionManager) CreateChannel(typeChan uint64) (channel *amqp.Channel, err error) {
	if typeChan == Producer {
		channel, err = p.prodConn.Channel()
	} else if typeChan == Consumer {
		channel, err = p.consConn.Channel()
	}
	return
}

//GetChannel returns the channel that is stored inside the map in the manager.
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

//InitChannel initialize channel with fn and add it to the map to recover or reinit.
func (p *ConnectionManager) InitChannel(fn InitializeChannel, args InitArgs) (err error) {
	if args.Channel == nil {
		err = ErrNilArg
		return
	}
	if p.isNotValidChannelArgs(args) {
		err = ErrInvalidArgs
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err = p.initChannel(fn, args)
	return
}

//Close close the connection and channel.
func (p *ConnectionManager) Close() (err error) {
	atomic.StoreUint32(&p.isClosed, Closed)
	err = p.prodConn.Close()
	err = p.consConn.Close()
	return
}

//IsClosed defines the state of connection in manager.
func (p *ConnectionManager) IsClosed() bool {
	return atomic.LoadUint32(&p.isClosed) == Closed
}

//connect creates new connection and add close notifier to trigger reinit or reconnect.
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
	return p.isNotValidKey(args.Key) || p.isNotValidTypeChan(args.TypeChan)
}

func (p *ConnectionManager) isNotValidTypeChan(typeChan uint64) bool {
	return typeChan == 0 || typeChan > 2
}

func (p *ConnectionManager) isNotValidKey(key string) bool {
	return key == ""
}

//reconnect does the reconnection of the rabbitmq manager.
func (p *ConnectionManager) reconnect() {
	var wg *sync.WaitGroup

	wg.Add(1)
	//Reconnecting for producers.
	go func() {
		defer wg.Done()
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
	wg.Add(1)
	//Reconnecting for consumers.
	go func() {
		defer wg.Done()
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
	wg.Wait()
	return
}

func (p *ConnectionManager) reinitConsumer() (err error) {
	var (
		newChan *amqp.Channel
	)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for key, customChannel := range p.consumer {
		customChannel.innerChan.Close()
		newChan, err = p.consConn.Channel()
		if err != nil {
			return
		}
		argsChan := InitArgs{
			Key:      key,
			TypeChan: Consumer,
			Channel:  newChan,
		}
		err = p.initChannel(customChannel.fn, argsChan)
		if err != nil {
			return
		}
	}
	return
}

func (p *ConnectionManager) reinitProducer() (err error) {
	var (
		newChan *amqp.Channel
	)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for key, customChannel := range p.producer {
		customChannel.innerChan.Close()
		newChan, err = p.prodConn.Channel()
		if err != nil {
			return
		}
		argsChan := InitArgs{
			Key:      key,
			TypeChan: Producer,
			Channel:  newChan,
		}
		err = p.initChannel(customChannel.fn, argsChan)
		if err != nil {
			return
		}
	}
	return
}
