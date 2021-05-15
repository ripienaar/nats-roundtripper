package nrt

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
)

// implements http.ResponseWriter
type response struct {
	gotStatus bool
	resp      *http.Response
	mu        sync.Mutex
}

func (r *NatsRoundTripper) ListenAndServ(ctx context.Context, hostname string, handler http.Handler) error {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	hostname = strings.ToLower(hostname)

	if !strings.HasSuffix(hostname, ".nats") {
		return fmt.Errorf("invalid hostname, hostnames should end in .nats")
	}

	subject := fmt.Sprintf("%s.requests.%s.*", r.opts.prefix, hostname)

	sub, err := r.opts.conn.QueueSubscribe(subject, "nrt", func(m *nats.Msg) {
		reqr := bufio.NewReader(bytes.NewBuffer(m.Data))
		req, err := http.ReadRequest(reqr)
		if err != nil {
			return
		}

		if req.Host != hostname {
			log.Printf("Ignoring request for host %s while serving %s", req.Host, hostname)
			return
		}

		log.Printf("[%s] Serving %s", hostname, req.URL.String())

		httpResp := &http.Response{Header: map[string][]string{}}
		handler.ServeHTTP(&response{resp: httpResp}, req)

		httpResp.Header.Set("NRT-ID", nuid.Next())
		httpResp.Header.Set("NRT-Connected-Server", r.opts.conn.ConnectedServerName())
		cid, err := r.opts.conn.GetClientID()
		if err == nil {
			httpResp.Header.Set("NRT-Client", strconv.Itoa(int(cid)))
		}
		cluster := r.opts.conn.ConnectedClusterName()
		if cluster != "" {
			httpResp.Header.Set("NRT-Connected-Cluster", r.opts.conn.ConnectedClusterName())
		}

		buf := bytes.NewBuffer([]byte{})
		err = httpResp.Write(buf)
		if err != nil {
			return
		}

		data := buf.Bytes()
		m.Respond(data)
	})
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Printf("Listening on %s", sub.Subject)

	<-ctx.Done()

	return nil
}

func (r *response) Header() http.Header {
	return r.resp.Header
}

func (r *response) Write(body []byte) (int, error) {
	bc := make([]byte, len(body))
	copy(bc, body)

	r.resp.Body = ioutil.NopCloser(bytes.NewBuffer(bc))

	if !r.gotStatus {
		r.WriteHeader(http.StatusOK)
		r.resp.Proto = "HTTP/1.1"
		r.resp.ProtoMinor = 1
		r.resp.ProtoMajor = 1
	}

	if r.resp.Header.Get("Content-Type") == "" {
		r.resp.Header.Set("Content-Type", http.DetectContentType(body))
	}

	r.resp.ContentLength = int64(len(body))

	return len(body), nil
}

func (r *response) WriteHeader(status int) {
	r.gotStatus = true
	r.resp.Status = http.StatusText(status)
	r.resp.StatusCode = status
}
