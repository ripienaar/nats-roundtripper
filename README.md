## NATS HTTP Round Tripper

This is a Golang `http.RoundTripper` that uses NATS as a transport.

### Why?

Using NATS as a transport for microservices has a number of advantages, you get a lot
of the benefits of a service mesh with no additional infrastructure.

 * Load balancing is free with no additional components like load balancers
 * Failover is free with no additional components like load balancers  
 * Geo failover and fallback is free with no additional components like global dns
 * Long running connections are used for microservices speeding up connection handling
 * NATS support canary deploys, gradual deploys and more (planned)

Moving to NATS as transport for microservices can be a big job, implementing a 
`http.RoundTripper` means your existing microservices can be hosted on nats with
minimal changes.

### Example?

See the [Weather Service](examples/weather)

The server is a normal HTTP server, no changes to your handlers:

```go
http.HandleFunc("/city", cityHandler)

// allow for gradual migration by listening on HTTP
log.Printf("Listening on 8080")
go http.ListenAndServe(":"+port, nil)

// Listen on NATS
nats, _ := nrt.New(nrt.WithNatsServer("nats.internal"))
go nats.ListenAndServ(context.Background(), "weather.nats", nil)

<-context.Background().Done()
```

The client is a normal HTTP client, see the bundled [nrtget](nrtget)

```
t, _ := nrt.New(nrt.WithNatsServer("nats.internal"))
c := &http.Client{Transport: t}

req, _ := http.NewRequest("GET", "http://weather.nats/city", bytes.NewBuffer([]byte(`{"city":"london"}`))))
resp, _ := c.Do(req)

defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)

fmt.Print(string(body))
```

### Design?

A server like `weather.nats` will subscribe in a queue group to the subject `nrt.requests.weather.nats.*`, clients
will publish a `GET` request to `nrt.requests.weather.nats.get`. 

The body being sent is the entire HTTP request, exactly as it came in the wire. On the server this is received and
turned into a standard `http.Request` and sent to the handler.

Response is simply a response containing the entire body with some headers added:

|Header|Meaning|
|------|-------|
|NRT-ID|A unique request ID set by the server|
|NRT-Connected-Server|The NATS server the microservice was connected to|
|NRT-Client|The Client ID of the server connection|
|NRT-Connected-Cluster|The cluster the server is connected to if clustered|

### Configuration ?

Several environment variables influence the behavior of the transport:

## Configuring

Configuration is via environment variables, below table details all environment variables:

|Variable|Description|
|--------|-----------|
|NATS_CONTEXT|A context to use as made by the `nats context save` command, see `NATS_ENV`|
|NATS_ENV|Set to `true` when no nats context is used, configure using the `NATS_*` variables instead|
|NATS_URL|Where the nats servers are|
|NATS_USER|User to connect as, also used as token when password is empty|
|NATS_PASSWORD|The password to connect with
|NATS_CREDS|Path to a nats creds file|
|NATS_NKEY|NKey to authenticate with|
|NATS_CERT|Path to a x509 certificate|
|NATS_KEY|Path to a x509 private key|
|NATS_CA|Path to a x509 certificate authority certificate chain|
|NATS_PREFIX|A subject to prefix to the hostname, defaults to `nrt`|

### Contact?

R.I.Pienaar / rip@devco.net / @ripienaar
