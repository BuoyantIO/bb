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
