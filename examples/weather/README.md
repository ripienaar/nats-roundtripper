## Weather API 

This is a sample microservice that uses the [OpenWeather](https://openweathermap.org/) API to serve current conditions in a city

It supports listening on HTTP and on NATS.

## Running

You need a OpenWeather API key, they are free just sign up.

You also need a NATS Context that configures the client for connecting to your network, here we'll use `demo.nats.io`

```
$ nats context save weather --server demo.nats.io
$ WEATHER_API=your.api.key \
  HTTP_PORT=8088 \
  NATS_CONTEXT=weather \
  ./weather
2021/05/15 10:20:20 Listening on :8088
2021/05/15 10:20:20 Listening on nrt.requests.weather.nats.* 
```

You can now test using `curl`:

```
$ curl http://localhost:8088/city -d '{"city":"london"}'
[
  {
    "id": 804,
    "main": "Clouds",
    "description": "overcast clouds",
    "icon": "04d"
  }
]
```

Or the bundled `nrtget`:

```
$ nrtget http://weather.nats/city '{"city":"london"}'
[
  {
    "id": 804,
    "main": "Clouds",
    "description": "overcast clouds",
    "icon": "04d"
  }
]
```
