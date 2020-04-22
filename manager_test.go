package amqpwrapper

import (
	"errors"
	"sync"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	type args struct {
		url    string
		config func() amqp.Config
	}
	tests := []struct {
		name        string
		args        args
		wantManager func() IConnectionManager
		wantErr     bool
	}{
		{
			name: "Initialize a new manager",
			args: args{
				url: uriDial,
				config: func() (config amqp.Config) {
					config = amqp.Config{
						Heartbeat: DefaultHeartbeat,
						Locale:    DefaultLocale,
					}
					return
				},
			},
			wantErr: false,
		},
		{
			name: "Empty Config",
			args: args{
				url: uriDial,
				config: func() (config amqp.Config) {
					config = amqp.Config{}
					return
				},
			},
			wantErr: false,
		},
		{
			name: "Empty url",
			args: args{
				url: "",
				config: func() (config amqp.Config) {
					config = amqp.Config{}
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid URL",
			args: args{
				url: "just uri test",
				config: func() (config amqp.Config) {
					config = amqp.Config{}
					return
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewManager(tt.args.url, tt.args.config())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestConnectionManager_CreateChannel(t *testing.T) {
	type args struct {
		typeChan uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create a consumer channel",
			args: args{
				typeChan: Consumer,
			},
			wantErr: false,
		},
		{
			name: "Create a producer channel",
			args: args{
				typeChan: Producer,
			},
			wantErr: false,
		},
		{
			name: "Invalid type channel",
			args: args{
				typeChan: 52,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewManager(uriDial, amqp.Config{})
			assert.Nilf(t, err, "Error in creating the connection manager: %s", err)
			_, err = p.CreateChannel(tt.args.typeChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectionManager.CreateChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestConnectionManager_GetChannel(t *testing.T) {
	type fields struct {
		manager func() IConnectionManager
	}
	type args struct {
		key      string
		typeChan uint64
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		checkChannel bool
	}{
		{
			name: "Get a producer channel",
			fields: fields{
				manager: func() IConnectionManager {
					newFn := func(amqpChan *amqp.Channel) (err error) {
						return
					}
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					amqpChan, err := mgr.CreateChannel(Producer)
					assert.Nilf(t, err, "Error in creating the new channel: %s", err)
					arg := InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
						Channel:  amqpChan,
					}
					err = mgr.InitChannel(newFn, arg)
					assert.Nilf(t, err, "Error in initializing the channel: %s", err)
					return mgr
				},
			},
			args: args{
				typeChan: Producer,
				key:      "Producer",
			},
			wantErr:      false,
			checkChannel: true,
		},
		{
			name: "Get a consumer channel",
			fields: fields{
				manager: func() IConnectionManager {
					newFn := func(amqpChan *amqp.Channel) (err error) {
						return
					}
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					amqpChan, err := mgr.CreateChannel(Consumer)
					assert.Nilf(t, err, "Error in creating the new channel: %s", err)
					arg := InitArgs{
						Key:      "Consumer",
						TypeChan: Consumer,
						Channel:  amqpChan,
					}
					err = mgr.InitChannel(newFn, arg)
					assert.Nilf(t, err, "Error in initializing the channel: %s", err)
					return mgr
				},
			},
			args: args{
				typeChan: Consumer,
				key:      "Consumer",
			},
			wantErr:      false,
			checkChannel: true,
		},
		{
			name: "Channel not initialized",
			fields: fields{
				manager: func() IConnectionManager {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					return mgr
				},
			},
			args: args{
				typeChan: Consumer,
				key:      "Consumer",
			},
			wantErr:      true,
			checkChannel: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChannel, err := tt.fields.manager().GetChannel(tt.args.key, tt.args.typeChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectionManager.GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkChannel {
				assert.NotNil(t, gotChannel, "Error channel must not be nil")
			} else {
				assert.Nil(t, gotChannel, "Channel must be nil")
			}
		})
	}
}

func TestConnectionManager_InitChannel(t *testing.T) {
	type fields struct {
		manager func() IConnectionManager
	}
	type args struct {
		fn   InitializeChannel
		args func() InitArgs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Init a new channel",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() (arg InitArgs) {
					arg = InitArgs{
						Key:      "Test-Consumer",
						Channel:  &amqp.Channel{},
						TypeChan: Consumer,
					}
					return
				},
			},
			wantErr: false,
		},
		{
			name: "Nil Channel",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() (arg InitArgs) {
					arg = InitArgs{
						Key:      "Test-Consumer",
						Channel:  nil,
						TypeChan: Consumer,
					}
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Empty Key",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() (arg InitArgs) {
					arg = InitArgs{
						Key:      "",
						Channel:  new(amqp.Channel),
						TypeChan: Consumer,
					}
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid type channel",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() (arg InitArgs) {
					arg = InitArgs{
						Key:      "Test-Consumer",
						Channel:  new(amqp.Channel),
						TypeChan: 52,
					}
					return
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fields.manager().InitChannel(tt.args.fn, tt.args.args()); (err != nil) != tt.wantErr {
				t.Errorf("ConnectionManager.InitChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionManager_Close(t *testing.T) {
	type fields struct {
		manager func() IConnectionManager
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Close the connection",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fields.manager().Close(); (err != nil) != tt.wantErr {
				t.Errorf("ConnectionManager.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionManager_IsClosed(t *testing.T) {
	type fields struct {
		manager func() IConnectionManager
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Check the close status of the connection (Status Open)",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					return
				},
			},
			want: false,
		},
		{
			name: "Check the close status of the connection (Status Closed)",
			fields: fields{
				manager: func() (connMgr IConnectionManager) {
					mgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error in initiating new manager: %s", err)
					connMgr = mgr
					connMgr.Close()
					return
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.manager().IsClosed(); got != tt.want {
				t.Errorf("ConnectionManager.IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulateReconnectManager(t *testing.T) {
	config := amqp.Config{}
	mgr := &ConnectionManager{
		url:      uriDial,
		config:   config,
		producer: make(map[string]*Channel, 0),
		consumer: make(map[string]*Channel, 0),
		mutex:    new(sync.RWMutex),
		wg:       new(sync.WaitGroup),
	}
	confirmChan := make(chan string, 1)
	mgr.connect()
	go mgr.reconnect()
	go func() {
		mgr.consConn.Close()
		confirmChan <- "Consumer connection closed"
		mgr.prodConn.Close()
		confirmChan <- "Producer connection closed"
		close(confirmChan)
	}()
	for text := range confirmChan {
		t.Logf("Connection close confirm: %s", text)
	}
}

func TestReinitConsumerProducer(t *testing.T) {
	t.Run("Simulate reinit success", func(t *testing.T) {
		config := amqp.Config{}
		mgr := &ConnectionManager{
			url:      uriDial,
			config:   config,
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		mgr.connect()
		newFn := func(amqpChan *amqp.Channel) (err error) {
			return
		}
		amqpChan, err := mgr.CreateChannel(Producer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg := InitArgs{
			Key:      "Producer",
			TypeChan: Producer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		amqpChan, err = mgr.CreateChannel(Consumer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg = InitArgs{
			Key:      "Consumer",
			TypeChan: Consumer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		//Simulate reinit in here
		err = mgr.reinitConsumer()
		assert.Nilf(t, err, "Error in reinitConsumer should be nil: %s", err)
		err = mgr.reinitProducer()
		assert.Nilf(t, err, "Error in reinitProducer should be nil: %s", err)
	})
	t.Run("Simulate reinit failed - consumer - closed connection", func(t *testing.T) {
		config := amqp.Config{}
		mgr := &ConnectionManager{
			url:      uriDial,
			config:   config,
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		mgr.connect()
		newFn := func(amqpChan *amqp.Channel) (err error) {
			return
		}
		amqpChan, err := mgr.CreateChannel(Producer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg := InitArgs{
			Key:      "Producer",
			TypeChan: Producer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		amqpChan, err = mgr.CreateChannel(Consumer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg = InitArgs{
			Key:      "Consumer",
			TypeChan: Consumer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		//Simulate reinit in here
		go func() {
			//Destroy consumer connection
			mgr.consConn.Close()
		}()
		err = mgr.reinitConsumer()
		assert.NotNil(t, err, "Error in reinitConsumer should be not nil.")
	})
	t.Run("Simulate reinit failed - consumer - init error", func(t *testing.T) {
		config := amqp.Config{}
		mgr := &ConnectionManager{
			url:      uriDial,
			config:   config,
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		mgr.connect()
		newFn := func(amqpChan *amqp.Channel) (err error) {
			return
		}
		amqpChan, err := mgr.CreateChannel(Producer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg := InitArgs{
			Key:      "Producer",
			TypeChan: Producer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		amqpChan, err = mgr.CreateChannel(Consumer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg = InitArgs{
			Key:      "Consumer",
			TypeChan: Consumer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		//Simulate reinit in here
		//Simulate init function error
		mgr.consumer["Consumer"].fn = func(amqpChan *amqp.Channel) (err error) {
			err = errors.New("Random Error")
			return
		}
		err = mgr.reinitConsumer()
		assert.NotNil(t, err, "Error in reinitConsumer should be not nil.")
	})
	t.Run("Simulate reinit failed - producer - closed connection", func(t *testing.T) {
		config := amqp.Config{}
		mgr := &ConnectionManager{
			url:      uriDial,
			config:   config,
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		mgr.connect()
		newFn := func(amqpChan *amqp.Channel) (err error) {
			return
		}
		amqpChan, err := mgr.CreateChannel(Producer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg := InitArgs{
			Key:      "Producer",
			TypeChan: Producer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		amqpChan, err = mgr.CreateChannel(Consumer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg = InitArgs{
			Key:      "Consumer",
			TypeChan: Consumer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		//Simulate reinit in here
		go func() {
			//Destroy consumer connection
			mgr.prodConn.Close()
		}()
		err = mgr.reinitProducer()
		assert.NotNil(t, err, "Error in reinitProducer should be not nil.")
	})
	t.Run("Simulate reinit failed - producer - init error", func(t *testing.T) {
		config := amqp.Config{}
		mgr := &ConnectionManager{
			url:      uriDial,
			config:   config,
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		mgr.connect()
		newFn := func(amqpChan *amqp.Channel) (err error) {
			return
		}
		amqpChan, err := mgr.CreateChannel(Producer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg := InitArgs{
			Key:      "Producer",
			TypeChan: Producer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		amqpChan, err = mgr.CreateChannel(Consumer)
		assert.Nilf(t, err, "Error in creating the new channel: %s", err)
		arg = InitArgs{
			Key:      "Consumer",
			TypeChan: Consumer,
			Channel:  amqpChan,
		}
		err = mgr.InitChannel(newFn, arg)
		assert.Nilf(t, err, "Error in initializing the channel: %s", err)
		//Simulate reinit in here
		//Simulate init function error
		mgr.producer["Producer"].fn = func(amqpChan *amqp.Channel) (err error) {
			err = errors.New("Random Error")
			return
		}
		err = mgr.reinitProducer()
		assert.NotNil(t, err, "Error in reinitProducer should be not nil.")
	})
}

func TestConnectError(t *testing.T) {
	t.Run("Connect producer error", func(t *testing.T) {
		mgr := &ConnectionManager{
			url:      "",
			config:   amqp.Config{},
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		err := mgr.connectProducer()
		assert.NotNil(t, err, "connectProducer error should not be nil")
	})
	t.Run("Connect consumer error", func(t *testing.T) {
		mgr := &ConnectionManager{
			url:      "",
			config:   amqp.Config{},
			producer: make(map[string]*Channel, 0),
			consumer: make(map[string]*Channel, 0),
			mutex:    new(sync.RWMutex),
			wg:       new(sync.WaitGroup),
		}
		err := mgr.connectConsumer()
		assert.NotNil(t, err, "connectConsumer error should not be nil")
	})
}

func TestConnectionManager_InitChannelAndGet(t *testing.T) {
	type args struct {
		fn   InitializeChannel
		args func() InitArgs
	}
	tests := []struct {
		name    string
		p       func() IConnectionManager
		args    args
		wantErr bool
	}{
		{
			name: "Function channel is nil",
			p: func() IConnectionManager {
				mgr, err := NewManager(uriDial, amqp.Config{})
				assert.Nilf(t, err, "Error in initiating new manager: %s", err)
				return mgr
			},
			args: args{
				args: func() InitArgs {
					return InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "Channel type is invalid",
			p: func() IConnectionManager {
				mgr, err := NewManager(uriDial, amqp.Config{})
				assert.Nilf(t, err, "Error in initiating new manager: %s", err)
				return mgr
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() InitArgs {
					return InitArgs{
						Key:      "Producer",
						TypeChan: 500,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "Connection is already closed",
			p: func() IConnectionManager {
				mgr, err := NewManager(uriDial, amqp.Config{})
				assert.Nilf(t, err, "Error in initiating new manager: %s", err)
				defer mgr.Close()
				return mgr
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() InitArgs {
					return InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "Init channel returns error",
			p: func() IConnectionManager {
				mgr, err := NewManager(uriDial, amqp.Config{})
				assert.Nilf(t, err, "Error in initiating new manager: %s", err)
				return mgr
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					err = errors.New("RANDOM")
					return
				},
				args: func() InitArgs {
					return InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "Initialize channel with nil channel",
			p: func() IConnectionManager {
				mgr, err := NewManager(uriDial, amqp.Config{})
				assert.Nilf(t, err, "Error in initiating new manager: %s", err)
				return mgr
			},
			args: args{
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				args: func() InitArgs {
					return InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
					}
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChannel, err := tt.p().InitChannelAndGet(tt.args.fn, tt.args.args())
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectionManager.InitChannelAndGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Nil(t, gotChannel, "The channel must be nil value")
			} else {
				assert.NotNil(t, gotChannel, "The channel must not be nil")
			}
		})
	}
}
