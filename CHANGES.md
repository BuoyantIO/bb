## In the next release...

* Better handling of contexts/deadlines/timeouts

## v0.0.5

* Fix gRPC clients to honor `downstream-timeout`.
* Remove `grpc-downstream-authority` flag in favor of `grpc-proxy` flag.
  If `grpc-proxy` is set, the target URL will be set by `grpc-proxy` and the
  `:authority` header will be set by `grpc-downstream-server`. If `grpc-proxy`
  is not set, the target URL will be set by `grpc-downstream-server`.
* Change `log-level` default from `debug` to `info`.
* Modify gRPC Client IDs to include `grpc-proxy` and `grpc-downstream-server`.
* Additional debug logging around broadcast requests.
* Bump `go-grpc` to `1.15.0`, `golang/protobuf` to `v1.2.0`, `logrus` to
  `v1.0.6`.
* Bump Docker build Golang version to `1.11.0`.

## v0.0.4

* Introduce `grpc-downstream-authority` flag, to enable setting authority
  separately from `grpc-downstream-server`.

## v0.0.3

* Update docs and example files to latest release.

## v0.0.2

* Introduce `terminate-after` flag, which instructs the process to shutdown
  after a specified number of requests.
* Introduce graceful shutdown. Upon receiving a shutdown message via SIGTERM, or
  via `terminate-after`, call shutdown on each server, allowing requests to
  drain.

## v0.0.1

bb 0.0.1 is the first public release of bb

* This release supports HTTP 1.1 and gRPC.
* Available strategies are: broadcast channel, point-to-point channel, terminus, and HTTP egress
* Allows users tio define a percentage of requests that should fail and a duration to wait for
  before processing requests.
* This release has been tested locally on Mac OS and on both Google Kubernetes Engine and
  Minikube.
