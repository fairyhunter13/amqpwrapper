package amqpwrapper

import (
	"errors"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestNewChannelManager(t *testing.T) {
	type args struct {
		key      string
		typeChan uint64
		fn       InitializeChannel
		connMgr  func() IConnectionManager
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create and init channel",
			args: args{
				key:      "Producer",
				typeChan: Producer,
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
			},
			wantErr: false,
		},
		{
			name: "Nil connection manager",
			args: args{
				key:      "Producer",
				typeChan: Producer,
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Key is not valid",
			args: args{
				key:      "",
				typeChan: Producer,
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Channel type is not valid",
			args: args{
				key:      "Producer",
				typeChan: 52,
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Create channel get error",
			args: args{
				key:      "Producer",
				typeChan: Producer,
				fn: func(amqpChan *amqp.Channel) (err error) {
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					mockConn := &MockIConnectionManager{}
					mockConn.On("CreateChannel", Producer).Return(nil, errors.New("Connection or random error"))
					connMgr = mockConn
					return
				},
			},
			wantErr: true,
		},
		{
			name: "Create channel get error",
			args: args{
				key:      "Producer",
				typeChan: Producer,
				fn: func(amqpChan *amqp.Channel) (err error) {
					err = errors.New("random error")
					return
				},
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewChannelManager(tt.args.key, tt.args.typeChan, tt.args.fn, tt.args.connMgr())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewChannelManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestChannelManager_Publish(t *testing.T) {
	type fields struct {
		connMgr  func() IConnectionManager
		key      string
		typeChan uint64
	}
	type args struct {
		exchange  string
		key       string
		mandatory bool
		immediate bool
		msg       amqp.Publishing
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Publish a message to mq",
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					amqpChan, err := connMgr.CreateChannel(Producer)
					assert.Nilf(t, err, "Error should be nil: %s", err)
					args := InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
						Channel:  amqpChan,
					}
					err = connMgr.InitChannel(func(*amqp.Channel) error {
						return nil
					}, args)
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
				key:      "Producer",
				typeChan: Producer,
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "Get channel failed",
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
				key:      "Producer",
				typeChan: Producer,
			},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ChannelManager{
				connMgr:  tt.fields.connMgr(),
				key:      tt.fields.key,
				typeChan: tt.fields.typeChan,
			}
			if err := m.Publish(tt.args.exchange, tt.args.key, tt.args.mandatory, tt.args.immediate, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("ChannelManager.Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChannelManager_Consume(t *testing.T) {
	type fields struct {
		connMgr  func() IConnectionManager
		key      string
		typeChan uint64
	}
	type args struct {
		queue     string
		consumer  string
		autoAck   bool
		exclusive bool
		noLocal   bool
		noWait    bool
		args      amqp.Table
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Consume a message from mq",
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					amqpChan, err := connMgr.CreateChannel(Producer)
					assert.Nilf(t, err, "Error should be nil: %s", err)
					args := InitArgs{
						Key:      "Producer",
						TypeChan: Producer,
						Channel:  amqpChan,
					}
					err = connMgr.InitChannel(func(amqpChan *amqp.Channel) (err error) {
						_, err = amqpChan.QueueDeclare("test", true, false, false, false, nil)
						return
					}, args)
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
				key:      "Producer",
				typeChan: Producer,
			},
			args: args{
				queue: "test",
			},
			wantErr: false,
		},
		{
			name: "Channel not found in the connection manager",
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					connMgr, err := NewManager(uriDial, amqp.Config{})
					assert.Nilf(t, err, "Error should be nil: %s", err)
					return
				},
				key:      "Producer",
				typeChan: Producer,
			},
			args: args{
				queue: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ChannelManager{
				connMgr:  tt.fields.connMgr(),
				key:      tt.fields.key,
				typeChan: tt.fields.typeChan,
			}
			_, err := m.Consume(tt.args.queue, tt.args.consumer, tt.args.autoAck, tt.args.exclusive, tt.args.noLocal, tt.args.noWait, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChannelManager.Consume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestChannelManager_IsClosed(t *testing.T) {
	type fields struct {
		connMgr  func() IConnectionManager
		key      string
		typeChan uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Channel manager is closed",
			want: true,
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					mgr := new(MockIConnectionManager)
					mgr.On("IsClosed").Return(true)
					connMgr = mgr
					return
				},
			},
		},
		{
			name: "Channel manager is not closed",
			want: false,
			fields: fields{
				connMgr: func() (connMgr IConnectionManager) {
					mgr := new(MockIConnectionManager)
					mgr.On("IsClosed").Return(false)
					connMgr = mgr
					return
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ChannelManager{
				connMgr:  tt.fields.connMgr(),
				key:      tt.fields.key,
				typeChan: tt.fields.typeChan,
			}
			if got := m.IsClosed(); got != tt.want {
				t.Errorf("ChannelManager.IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}
