package nrt

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Option configures the NatsRoundTripper
type Option func(opts *nrtOpts) error

type nrtOpts struct {
	conn        *nats.Conn
	natsURL     string
	connOpts    []nats.Option
	contextName string
	prefix      string
	timeout     time.Duration
}

// WithNatsOptions creates a nats connection with these options, if given with a context will be appended to those options set by the context
func WithNatsOptions(opts ...nats.Option) Option {
	return func(o *nrtOpts) error {
		o.connOpts = opts
		return nil
	}
}

// WithContextNamed uses a context as created by nats cli for connection, can be augmented by WithNatsOptions()
func WithContextNamed(n string) Option {
	return func(o *nrtOpts) error {
		o.contextName = n
		return nil
	}
}

// WithConnection supplies a prepared nats connection, not compatible with WithNatsOptions() and WithContextNamed()
func WithConnection(nc *nats.Conn) Option {
	return func(o *nrtOpts) error {
		o.conn = nc
		return nil
	}
}

// WithPrefix sets a specific subject prefix, defaults to `nrtt`
func WithPrefix(s string) Option {
	return func(o *nrtOpts) error {
		o.prefix = s
		return nil
	}
}

// WithTimeout sets the timeout for various operations, defaults to 30s
func WithTimeout(t time.Duration) Option {
	return func(o *nrtOpts) error {
		o.timeout = t
		return nil
	}
}

// WithNatsServer sets the url(s) to the nats servers, required when no context or connection is given
func WithNatsServer(u string) Option {
	return func(o *nrtOpts) error {
		o.natsURL = u
		return nil
	}
}
