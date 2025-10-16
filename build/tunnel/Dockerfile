FROM golang:1.25 AS build

WORKDIR /tunnel

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /tunnel ./cmd/tunnel/tunnel.go
RUN groupadd -r tunnel && useradd --no-log-init -r -g tunnel tunnel

FROM scratch AS release

WORKDIR /tunnel

COPY --from=build /tunnel/tunnel /tunnel/tunnel
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

USER tunnel:tunnel

ENTRYPOINT ["/tunnel/tunnel"]

