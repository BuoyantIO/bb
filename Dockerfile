# cache go modules in a separate image
FROM --platform=$BUILDPLATFORM golang:1.22.3-alpine as go-deps
WORKDIR /bb-build
COPY go.mod go.sum main.go ./
COPY cmd cmd
COPY gen gen
COPY protocols protocols
COPY service service
COPY strategies strategies
RUN go mod vendor

# build the bb binary
FROM --platform=$BUILDPLATFORM go-deps as golang
WORKDIR /bb-build
RUN CGO_ENABLED=0 go build -o /out/bb -mod=vendor .

# package a runtime image
FROM scratch
LABEL org.opencontainers.image.source=https://github.com/buoyantio/bb
COPY --from=golang /out/bb /out/bb
COPY --from=golang /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/out/bb"]
