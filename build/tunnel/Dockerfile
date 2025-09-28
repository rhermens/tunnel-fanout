FROM golang:1.25

WORKDIR /tunnel

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /tunnel ./cmd/tunnel/tunnel.go

CMD ["/tunnel"]

