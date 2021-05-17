## Caddy Middleware Plugin

This is a experimental plugin for Caddy V2 that allow you to proxy HTTP requests to a NATS backed set of backends.

```
$ cat Caddyfile
{
    order nats last
}

http://:9090 {
        nats /* weather.nats 
}
```

Now any request to port 9090 will forward to the NATS hosted service cluster.

```
$ curl http://localhost:9090/city -d '{"city":"london"}' -v
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 9090 (#0)
> POST /city HTTP/1.1
> Host: localhost:9090
> User-Agent: curl/7.64.1
> Accept: */*
> Content-Length: 17
> Content-Type: application/x-www-form-urlencoded
>
* upload completely sent off: 17 out of 17 bytes
< HTTP/1.1 200 OK
< Server: Caddy
< Date: Mon, 17 May 2021 14:37:02 GMT
< Content-Length: 293
< Content-Type: text/plain; charset=utf-8
<
HTTP/1.1 200 OK
Content-Length: 104
Content-Type: text/plain; charset=utf-8
Nrt-Client: 4019
Nrt-Connected-Cluster: lon
Nrt-Connected-Server: n2-lon
Nrt-Id: VYOnC1jYhwUAR15DsIecFH

[
  {
    "id": 804,
    "main": "Clouds",
    "description": "overcast clouds",
    "icon": "04d"
  }
]
```
