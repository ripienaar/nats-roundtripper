package caddy_nrt

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nats-io/jsm.go/natscontext"
	nrt "github.com/ripienaar/nats-roundtripper"
	"go.uber.org/zap"
)

type Nats struct {
	Context string `json:"context"`
	Service string `json:"service"`

	log *zap.Logger
	rt  *nrt.NatsRoundTripper
	mu  sync.Mutex
}

func init() {
	caddy.RegisterModule(&Nats{})
	httpcaddyfile.RegisterHandlerDirective("nats", parseCaddyfile)
}

func (n *Nats) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) < 1 {
			return fmt.Errorf("a service name is required")
		}

		n.Service = args[0]
	}

	return nil
}

func (n *Nats) Validate() error {
	if n.Context == "" {
		return fmt.Errorf("a context is required, set environment variable NATS_CONTEXT")
	}
	if n.Service == "" {
		return fmt.Errorf("a service is required")
	}

	if !natscontext.IsKnown(n.Context) {
		return fmt.Errorf("unknown context %s", n.Context)
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var n Nats
	err := n.UnmarshalCaddyfile(h.Dispenser)
	n.Context = os.Getenv("NATS_CONTEXT")
	return &n, err
}

func (n *Nats) Provision(ctx caddy.Context) error {
	n.log = ctx.Logger(n)

	n.log.Info(fmt.Sprintf("Provisioning NATS transport for service %s", n.Service))

	return nil
}

func (n *Nats) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	n.mu.Lock()
	if n.rt == nil {
		var err error
		n.log.Info(fmt.Sprintf("Connecting to NATS using context %s", n.Context))
		n.rt, err = nrt.New()
		if err != nil {
			n.mu.Unlock()
			return err
		}
	}
	n.mu.Unlock()

	nreq := r.Clone(context.Background())
	nreq.Host = n.Service
	nreq.URL.Host = n.Service

	natsResp, err := n.rt.RoundTrip(nreq)
	if err != nil {
		return err
	}

	natsResp.Write(w)

	return next.ServeHTTP(w, r)
}

func (n Nats) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.nats",
		New: func() caddy.Module { return &Nats{} },
	}
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Nats)(nil)
	_ caddyhttp.MiddlewareHandler = (*Nats)(nil)
	_ caddy.Validator             = (*Nats)(nil)
	_ caddyfile.Unmarshaler       = (*Nats)(nil)
)
