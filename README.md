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
|NRT-HostName|Added to request messages to help construct the request in the server|
|NRT-Method|The HTTP method added to request messages to help construct the request in the server|
|NRT-Path|The HTTP request path added to request messages to help construct the request in the server|
|NRT-ID|A unique request ID set by the server|
|NRT-Connected-Server|The NATS server the microservice was connected to|
|NRT-Client|The Client ID of the server connection|
|NRT-Connected-Cluster|The cluster the server is connected to if clustered|

### Contact?

R.I.Pienaar / rip@devco.net / @ripienaar
