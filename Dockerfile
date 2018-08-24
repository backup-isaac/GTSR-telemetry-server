FROM golang:1.9.2

WORKDIR /go/src/telemetry-server/
COPY . .

RUN go get -v -t ./...
CMD ["go", "run", "main.go"]