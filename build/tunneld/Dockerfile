FROM golang:1.25

WORKDIR /tunneld

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /tunneld ./cmd/tunneld/tunneld.go

CMD ["/tunneld"]

