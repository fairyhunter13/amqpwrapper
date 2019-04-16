package rabbitmq

import (
	"reflect"
	"sync"
	"testing"

	"github.com/streadway/amqp"
)

func TestManager_GetChannel(t *testing.T) {
	type fields struct {
		producer func() *sync.Map
		consumer func() *sync.Map
		prodConn *amqp.Connection
		consConn *amqp.Connection
		isClosed uint32
	}
	type args struct {
		key      string
		typeChan int64
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantChannel func() *amqp.Channel
		wantErr     bool
	}{
		{
			name: "Get producer channel from normal condition",
			fields: fields{
				producer: func() (m *sync.Map) {
					m = new(sync.Map)
					m.Store("tester", &Channel{
						innerChan: new(amqp.Channel),
					})
					return
				},
				consumer: func() (m *sync.Map) {
					m = new(sync.Map)
					return
				},
			},
			args: args{
				key:      "tester",
				typeChan: Producer,
			},
			wantErr: false,
			wantChannel: func() (channel *amqp.Channel) {
				channel = new(amqp.Channel)
				return
			},
		},
		{
			name: "Get consumer channel from normal condition",
			fields: fields{
				producer: func() (m *sync.Map) {
					m = new(sync.Map)
					return
				},
				consumer: func() (m *sync.Map) {
					m = new(sync.Map)
					m.Store("tester", &Channel{
						innerChan: new(amqp.Channel),
					})
					return
				},
			},
			args: args{
				key:      "tester",
				typeChan: Consumer,
			},
			wantErr: false,
			wantChannel: func() (channel *amqp.Channel) {
				channel = new(amqp.Channel)
				return
			},
		},
		{
			name: "Channel not found in the map",
			fields: fields{
				producer: func() (m *sync.Map) {
					m = new(sync.Map)
					return
				},
				consumer: func() (m *sync.Map) {
					m = new(sync.Map)
					return
				},
			},
			args: args{
				key:      "tester",
				typeChan: Consumer,
			},
			wantErr: true,
			wantChannel: func() (channel *amqp.Channel) {
				return
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Manager{
				producer: tt.fields.producer(),
				consumer: tt.fields.consumer(),
				prodConn: tt.fields.prodConn,
				consConn: tt.fields.consConn,
				isClosed: tt.fields.isClosed,
			}
			gotChannel, err := p.GetChannel(tt.args.key, tt.args.typeChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotChannel, tt.wantChannel()) {
				t.Errorf("Manager.GetChannel() = %v, want %v", gotChannel, tt.wantChannel())
			}
		})
	}
}
