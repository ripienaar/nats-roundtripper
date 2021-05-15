package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/nats-io/jsm.go/natscontext"
	nrt "github.com/ripienaar/nats-roundtripper"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	c := &http.Client{Transport: nrt.Must(nrt.WithContextNamed(natscontext.SelectedContext()))}

	req, err := http.NewRequest("GET", os.Args[1], bytes.NewBuffer([]byte(os.Args[2])))
	panicIfErr(err)
	resp, err := c.Do(req)
	panicIfErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	panicIfErr(err)

	fmt.Print(string(body))
}
