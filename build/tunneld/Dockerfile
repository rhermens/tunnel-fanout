FROM golang:1.25 AS build

WORKDIR /tunneld

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /tunneld ./cmd/tunneld/tunneld.go

FROM scratch AS release

WORKDIR /tunneld

COPY --from=build /tunneld/tunneld /tunneld/tunneld

ENTRYPOINT ["/tunneld/tunneld"]
