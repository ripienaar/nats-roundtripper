## NATS HTTP Round Tripper

This is a Golang `http.RoundTripper` that uses NATS as a transport.

Included is a `http.RoundTripper` for clients, a server that uses normal HTTP Handlers and any existing http handler mux and a Caddy Server transport.

### Why?

Using NATS as a transport for microservices has a number of advantages, you get a lot
of the benefits of a service mesh with no additional infrastructure.

 * Load balancing is free with no additional components like load balancers
 * Failover is free with no additional components like load balancers  
 * Geo failover and fallback is free with no additional components like global dns
 * Long running connections are used for microservices speeding up connection handling
 * NATS support canary deploys, gradual deploys, traffic shaping or more
 * NATS, being subject and interest based, does not need any additional infrastructure for service discovery

Moving to NATS as transport for microservices can be a big job, implementing a 
`http.RoundTripper` means your existing microservices can be hosted on nats with
minimal changes.

You can see it in action [on YouTube](https://www.youtube.com/watch?v=qZ7fww7_s7A) doing geo failover.

### Example?

#### Server / Microservice

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

#### Client

The client is a normal HTTP client, see the bundled [nrtget](nrtget)

```go
t, _ := nrt.New(nrt.WithNatsServer("nats.internal"))
c := &http.Client{Transport: t}

req, _ := http.NewRequest("GET", "https://weather.nats/city", bytes.NewBuffer([]byte(`{"city":"london"}`))))
resp, _ := c.Do(req)

defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)

fmt.Print(string(body))
```

#### Proxy

I have a Caddy Server 2 extension that you can build using [xcaddy](https://github.com/caddyserver/xcaddy), using this you can bridge
existing HTTP traffic to a backend managed by NATS.

This extension is just an experiment, the Caddy 2 documentation is non existent so doing the right thing is quite hard. Caddy is really
attractive though as it's high speed, easy to extend and support many features like Lets Encrypt right out of the box.

You'd look to use something like this to bridge existing traffic to your backend, or to assist with migrating from a current REST based solution to a NATS one.


```
{
    order nats last
}

http://:9090 {
        nats /city weather.nats
}
```

Run this using the environment variables listed below like `NATS_CONTEXT` for configuration.

With this in place requests to `http://example.net:9090/city` will be sent to the NATS backend `weather.nats` microservice.

This has a number of advantages:

 * A long-running connection from Caddy to NATS is made significantly reducing the cost overhead of connection management, especially with TLS
 * Despite being a single long running connection load balancing and geo failover is supported
 * The connection is self managing and will reconnect automatically, supports NATS topology discovery and will keep working through scale ups and downs of the NATS infrastructure
 * Massive infrastructure reduction as you don't need any service discovery or LB level configuration to do traffic shaping etc

### Traffic Management

#### Load balancing

Without any additional effort the service will scale horizontally and vertically, simply starting more of the same code will cause the load to be randomly distributed.

#### Geo Failover

If instances of the service is started in multiple NATS Clustered connected into a Super Cluster then connections made in cluster "east" will automatically failover to cluster "west" and fall back to "east" when workers are available.

#### Canary Deploys

When running a service the `NATS_PREFIX` environment variable or the `WithPrefix()` option can adjust the subjects where the service listen.

Here we run the weather service on a `v2` subject:

```
$ WEATHER_API=your.api.key \
  HTTP_PORT=8088 \
  NATS_CONTEXT=weather \
  NATS_PREFIX=v2 \
  ./weather
```

Lets assume we have both a `v1` and `v2` running, and we wish to test 10% of the traffic against `v2`.  Configuring the NATS Server as follows will set this up:

```
mappings: {
    nrt.requests.weather.nats.*: [
        {destination: v1.requests.weather.nats.$1, weight: 90%}
        {destination: v2.requests.weather.nats.$1, weight: 10%}
    ]
}
```

With this in place 10% of the traffic will be routed to v2 of the microservice. These settings can be adjusted on the fly to migrate traffic onto v2 gradually.


#### Fault Injection

Building on the mappings above we can arrange for 10% of the traffic to be lost:

```
mappings: {
    nrt.requests.weather.nats.*: [
        {destination: v1.requests.weather.nats.$1, weight: 90%}
        {destination: v1.discard.weather.nats.$1, weight: 10%}
    ]
}
```

Here we ensure no services listens on `v1.discard...` and those 10% of traffic will be lost.

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

Several environment variables influence the behavior of the transport, most have matching options to configure the connections:

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
