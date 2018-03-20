## v0.0.1

bb 0.0.1 is the first public release of bb

* This release supports HTTP 1.1 and gRPC.
* Available strategies are: broadcast channel, point-to-point channel, terminus, and HTTP egress
* Allows users tio define a percentage of requests that should fail and a duration to wait for 
  before processing requests.
* This release has been tested locally on Mac OS and on both Google Kubernetes Engine and 
  Minikube.
