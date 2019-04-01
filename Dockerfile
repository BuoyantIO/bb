FROM golang:1.11.6-stretch as golang
WORKDIR /go/src/github.com/buoyantio/bb
ADD .  /go/src/github.com/buoyantio/bb

RUN mkdir -p /out
RUN ./bin/dep ensure
RUN go build -o /out/bb .

FROM gcr.io/linkerd-io/base:2019-02-19.01
RUN apt-get update
RUN apt-get install -y ca-certificates
COPY --from=golang /out /out
ENTRYPOINT ["/out/bb"]
