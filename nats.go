package nrt

import (
	"time"

	"github.com/nats-io/jsm.go/natscontext"
	"github.com/nats-io/nats.go"
)

type NatsRoundTripper struct {
	opts *nrtOpts
}

// New creates a new http.RoundTripper that uses NATS as a transport
func New(opts ...Option) (*NatsRoundTripper, error) {
	nrt := &NatsRoundTripper{
		opts: &nrtOpts{
			prefix:  "nrt",
			timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		err := opt(nrt.opts)
		if err != nil {
			return nil, err
		}
	}

	err := nrt.connect()
	if err != nil {
		return nil, err
	}

	return nrt, nil
}

// Must creates a new http.RoundTripper that uses NATS as a transport, panic on error
func Must(opts ...Option) *NatsRoundTripper {
	nrt, err := New(opts...)
	if err != nil {
		panic(err)
	}

	return nrt
}

func (r *NatsRoundTripper) connect() error {
	if r.opts.conn != nil {
		return nil
	}

	var (
		opts []nats.Option
		err  error
		url  string
	)

	if r.opts.contextName != "" {
		nctx, err := natscontext.New(r.opts.contextName, true)
		if err != nil {
			return err
		}

		opts, err = nctx.NATSOptions()
		if err != nil {
			return err
		}

		url = nctx.ServerURL()
	}

	opts = append(opts, r.opts.connOpts...)

	if r.opts.natsURL != "" {
		url = r.opts.natsURL
	}

	// TODO: error callbacks etc

	r.opts.conn, err = nats.Connect(url, opts...)

	return err
}
