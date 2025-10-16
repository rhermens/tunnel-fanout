FROM golang:1.25 AS build

WORKDIR /tunneld

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /tunneld ./cmd/tunneld/tunneld.go
RUN groupadd -r tunneld && useradd --no-log-init -r -g tunneld tunneld

FROM scratch AS release

WORKDIR /tunneld

COPY --from=build /tunneld/tunneld /tunneld/tunneld
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

ENTRYPOINT ["/tunneld/tunneld"]
