package nrt

import (
	"os"
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
			prefix:      "nrt",
			timeout:     30 * time.Second,
			contextName: os.Getenv("NATS_CONTEXT"),
		},
	}

	if prefix := os.Getenv("NATS_PREFIX"); prefix != "" {
		nrt.opts.prefix = prefix
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

	// loads the context, if you dont have a context you can set all the connection
	// options via os environment, set NATS_ENV=true and then supply the variables
	// NATS_URL, NATS_USER etc
	envContext := os.Getenv("NATS_ENV") == "true"
	if r.opts.contextName != "" || envContext {
		nctx, err := natscontext.New(r.opts.contextName, !envContext,
			natscontext.WithServerURL(os.Getenv("NATS_URL")),
			natscontext.WithUser(os.Getenv("NATS_USER")),
			natscontext.WithPassword(os.Getenv("NATS_PASSWORD")),
			natscontext.WithCreds(os.Getenv("NATS_CREDS")),
			natscontext.WithNKey(os.Getenv("NATS_NKEY")),
			natscontext.WithCertificate(os.Getenv("NATS_CERT")),
			natscontext.WithKey(os.Getenv("NATS_KEY")),
			natscontext.WithCA(os.Getenv("NATS_CA")),
		)
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
	opts = append(opts, nats.MaxReconnects(-1))

	if r.opts.natsURL != "" {
		url = r.opts.natsURL
	}

	r.opts.conn, err = nats.Connect(url, opts...)

	return err
}
