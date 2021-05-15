package nrt

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/nats-io/nats.go"
)

func (r *NatsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	hostname := strings.ToLower(req.URL.Hostname())
	if !strings.HasSuffix(hostname, ".nats") {
		return nil, fmt.Errorf("invalid hostname, hostnames should end in .nats")
	}

	subject := fmt.Sprintf("%s.requests.%s.%s", r.opts.prefix, hostname, strings.ToLower(req.Method))

	body := bytes.NewBuffer([]byte{})
	err := req.Write(body)
	if err != nil {
		return nil, err
	}

	msg := nats.Msg{Data: body.Bytes(), Header: map[string][]string{}}
	msg.Header.Set("NRT-Method", req.Method)
	msg.Header.Set("NRT-Path", req.URL.Path)
	msg.Header.Set("NRT-HostName", hostname)
	msg.Subject = subject

	res, err := r.opts.conn.RequestMsg(&msg, r.opts.timeout)
	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(res.Data)), req)
}
