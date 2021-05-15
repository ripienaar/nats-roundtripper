package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	owm "github.com/briandowns/openweathermap"
	nrt "github.com/ripienaar/nats-roundtripper"
)

var apiKey = os.Getenv("WEATHER_API")

type CityRequest struct {
	City string `json:"city"`
}

func main() {
	// weather api is needed
	if apiKey == "" {
		panic("Please set WEATHER_API")
	}

	// standard http handler unchanged
	http.HandleFunc("/city", cityHandler)

	// set HTTP_PORT to listen on standard HTTP
	port := os.Getenv("HTTP_PORT")
	if port != "" {
		pi, err := strconv.Atoi(port)
		if err != nil {
			panic(err)
		}

		log.Printf("Listening on :%d", pi)
		go http.ListenAndServe(":"+port, nil)
	}

	err := nrt.Must().ListenAndServ(context.Background(), "weather.nats", nil)
	if err != nil {
		panic(err)
	}
}

func cityHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Could not read body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not read request"))
		return
	}

	creq := &CityRequest{}
	err = json.Unmarshal(body, &creq)
	if err != nil {
		log.Printf("Could not parse body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not parse body"))
		return
	}

	if creq.City == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		return
	}

	weather, err := owm.NewCurrent("C", "EN", apiKey)
	if err != nil {
		log.Printf("Could not initialize API: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("weather api error"))
		return
	}

	err = weather.CurrentByName(creq.City)
	if err != nil {
		log.Printf("Lookup failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not request weather"))
		return
	}

	rj, err := json.MarshalIndent(weather.Weather, "", "  ")
	if err != nil {
		log.Printf("Response create failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not process API result"))
		return
	}

	w.Write(rj)
}
