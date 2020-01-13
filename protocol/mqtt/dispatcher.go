package mqtt

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/countstarlight/homo/logger"
	"github.com/countstarlight/homo/utils"
	"github.com/jpillora/backoff"
	"go.uber.org/zap"
	"time"
)

// Dispatcher dispatcher of mqtt client
type Dispatcher struct {
	config  ClientInfo
	channel chan packet.Generic
	backoff *backoff.Backoff
	tomb    utils.Tomb
	log     *zap.SugaredLogger
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(cc ClientInfo, log *zap.SugaredLogger) *Dispatcher {
	if log == nil {
		log = logger.S
	}
	return &Dispatcher{
		config:  cc,
		channel: make(chan packet.Generic, cc.BufferSize),
		backoff: &backoff.Backoff{
			Min:    time.Millisecond * 500,
			Max:    cc.Interval,
			Factor: 2,
		},
		log: log.With("mqtt", "dispatcher").With("cid", cc.ClientID),
	}
}

// Start starts dispatcher
func (d *Dispatcher) Start(h Handler) error {
	return d.tomb.Go(func() error {
		return d.supervisor(h)
	})
}

// Close closes dispatcher
func (d *Dispatcher) Close() error {
	d.tomb.Kill(nil)
	return d.tomb.Wait()
}

// Supervisor the supervised reconnect loop
func (d *Dispatcher) supervisor(handler Handler) error {
	first := true
	var dying bool
	var current packet.Generic

	for {
		if first {
			// no delay on first attempt
			first = false
		} else {
			// get backoff duration
			next := d.backoff.Duration()

			d.log.Debug("delay reconnect:", next)

			// sleep but return on Stop
			select {
			case <-time.After(next):
			case <-d.tomb.Dying():
				return nil
			}
		}

		d.log.Debug("next reconnect")

		client, err := NewClient(d.config, handler, d.log)
		if err != nil {
			d.log.Errorw("failed to create new client", zap.Error(err))
			continue
		}

		// run callback
		d.log.Debug("client online")

		// run dispatcher on client
		current, dying = d.dispatcher(client, current)

		// run callback
		d.log.Debug("client offline")

		// return goroutine if dying
		if dying {
			return nil
		}
	}
}

// reads from the queues and calls the current client
func (d *Dispatcher) dispatcher(client *Client, current packet.Generic) (packet.Generic, bool) {
	defer client.Close()

	if current != nil {
		err := client.Send(current)
		if err != nil {
			return current, false
		}
	}

	for {
		select {
		case pkt := <-d.channel:
			err := client.Send(pkt)
			if err != nil {
				return pkt, false
			}
		case <-client.Dying():
			return nil, false
		case <-d.tomb.Dying():
			return nil, true
		}
	}
}
