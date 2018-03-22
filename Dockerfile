FROM golang:1.10.0-stretch as golang
WORKDIR /go/src/github.com/buoyantio/bb
ADD .  /go/src/github.com/buoyantio/bb
RUN mkdir -p /out
RUN go build -o /out/bb .

FROM gcr.io/runconduit/base:2017-10-30.01
RUN apt-get update
RUN apt-get install -y ca-certificates
COPY --from=golang /out /out
ENTRYPOINT ["/out/bb"]