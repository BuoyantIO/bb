ARG BUILDPLATFORM=linux/amd64

FROM --platform=$BUILDPLATFORM golang:1.11.6-stretch as golang
WORKDIR /go/src/github.com/buoyantio/bb
ADD .  /go/src/github.com/buoyantio/bb

RUN mkdir -p /out
RUN ./bin/dep ensure
ARG TARGETARCH
RUN GOOS=linux GOARCH=$TARGETARCH go build -o /out/bb .

FROM debian:buster-20200514-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    dnsutils \
    iptables \
    jq \
    nghttp2 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# We still rely on old iptables-legacy syntax.
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy \
    && update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

COPY --from=golang /out /out
ENTRYPOINT ["/out/bb"]
