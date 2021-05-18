FROM goboring/golang:1.16.4b7

WORKDIR /go/src/app
COPY . .

RUN go install -v ./...

CMD ["weather"]
