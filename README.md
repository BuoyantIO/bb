[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/BuoyantIO/bb.svg?style=shield)](https://circleci.com/gh/BuoyantIO/bb)

# bb

Building Blocks or `bb` is a tool that can simulate many of the typical scenarios
of a cloud-native Service-Oriented Architecture based on microservices.

## Using `bb`
`bb` publishes a single container, `buoyantio/bb:v0.0.5`. Instances of this
container receive and return a simple message, described by the protobuf schema
[in this repository](api.proto). This known interface allows for `bb`
containers to be arranged in many different ways, just like building a structure
using LEGO blocks.

The best way to find out about `bb` features is by using the `--help` command and
checking out the self-contained documentation.

### Running locally
If you want to run `bb` locally, outside a Kubernetes cluster or Docker, you need
to build a binary for your environment. The usual way to go about this is using Go's
standard build tools. From this repository's root, run:

    $ mkdir -p target && go build -o target/bb .

This will create a `bb` binary in the `target` directory. You can run this on your
computer but, unless you use Linux, you won't be able to use this same binary on
Docker or Kubernetes.

As an example of how to use `bb` on your computer, let's create a simple two-service
setup. Open a terminal and type in the following:

    $  target/bb terminus --grpc-server-port 9090 --response-text BANANA

Then, on a second terminal window, type this:

    $ target/bb point-to-point-channel --grpc-downstream-server localhost:9090 --h1-server-port 8080

Now, on a third terminal, type this and you should see a similar response:

    $ curl localhost:8080
    {"requestUid":"in:http-sid:point-to-point-channel-grpc:-1-h1:8080-387107000","payload":"BANANA"}

The first command you typed created a gRPC server on port 9090 following the [terminus strategy](strategies/terminus.go),
which will return the payload defined as the `--response-text` argument for any valid request.

The second command creates an HTTP 1.1 server on port 8080 following the
[point-to-point channel strategy](strategies/point_to_point_channel.go), and has the gRPC server
previously defined as its downstream. It will receive any HTTP 1.1 request, convert it to gRPC,
and forward it to the downstream server. It will also get the gRPC response from the server, convert
it to JSON-over-HTTP, and return to its client.

## Running on Kubernetes
Although `bb` can be useful to test things locally as described above, its main use case is to create
complicated environments inside Kubernetes clusters.

To use `bb` with Kubernetes, the first step you need to take is to publish its Docker image to your
Kubernetes cluster. Here we will be using a local Minikube installation to demonstrate its use.

First, make sure that Minikube is running:

    $ minikube status
      minikube: Running
      cluster: Running
      kubectl: Correctly Configured: pointing to minikube-vm at 192.168.99.100

Now, make sure that your Docker environment variables are set to use Minikube as the Docker repository
for images:

    $ eval "$(minikube docker-env)"

You should then build a Docker image for `bb`:

    $  bin/docker-build-bb.sh
    Sending build context to Docker daemon  136.7MB
    Step 1/6 : FROM buoyantio/base:2017-10-30.01
     ---> 14aa74f25501
    Step 2/6 : RUN apt-get update
    [...]
    Removing intermediate container f4f571b01dd8
    Successfully built e6d76c5df612
    Successfully tagged buoyantio/bb:v0.0.5

A test run using the Docker CLI should return usage information and confirm that everything is ok:

    $ docker run buoyantio/bb:v0.0.5

    Building Blocks or `bb` is a tool that can simulate many of the typical scenarios of a cloud-native Service-Oriented Architecture based on microservices.

    Usage:
      bb [command]

    Available Commands:
      broadcast-channel      Forwards the request to all downstream services.
      help                   Help about any command
      http-egress            Receives a request, makes a HTTP(S) call to a specified URL and return the body of the response
      point-to-point-channel Forwards the request to one and only one downstream service.
      terminus               Receives the request and returns a pre-defined response

    Flags:
          --downstream-timeout duration      timeout to use when making downstream connections and requests. (default 1m0s)
          --fire-and-forget                  do not wait for a response when contacting downstream services.
          --grpc-downstream-server strings   list of servers (hostname:port) to send messages to using gRPC, can be repeated
          --grpc-proxy string                optional proxy to route gRPC requests
          --grpc-server-port int             port to bind a gRPC server to (default -1)
          --h1-downstream-server strings     list of servers (protocol://hostname:port) to send messages to using HTTP 1.1, can be repeated
          --h1-server-port int               port to bind a HTTP 1.1 server to (default -1)
      -h, --help                             help for bb
          --id string                        identifier for this container
          --log-level string                 log level, must be one of: panic, fatal, error, warn, info, debug (default "info")
          --percent-failure int              percentage of requests that this service will automatically fail
          --sleep-in-millis int              amount of milliseconds to wait before actually start processing a request
          --terminate-after int              terminate the process after this many requests

    Use "bb [command] --help" for more information about a command.

To build the exact same scenario we had above, but for Kubernetes, you can deploy the should have
a YAML configuration like the one in our [examples directory](examples). You can deploy it to your Kubernetes
cluster by running:

    $ kubectl apply -f examples/bb-readme/application.yaml

You can then use `curl`to query the service:

    $ curl `minikube -n bb-readme service bb-readme-gateway-svc  --url`
    {"requestUid":"in:http-sid:point-to-point-channel-grpc:-1-h1:8080-66349706","payload":"BANANA"}

## Build Multi-Arch Images with Buildx

    docker buildx build . \
      --platform linux/amd64,linux/arm64,linux/arm/v7 \
      --push \
      --tag buoyantio/bb:<version>
