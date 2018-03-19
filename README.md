# bb

Building Blocks or `bb` is a tool that can simulate many of the typical scenarios 
of a cloud-native Service-Oriented Architecture based on microservices.

## Using `bb`
`bb` publishes a single container, `gcr.io/runconduit/bb:v1`. Instances of this 
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

Them, on a second terminal window, type this:

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

    $  bin/docker-build-building-blocks.sh
    Sending build context to Docker daemon  136.7MB
    Step 1/6 : FROM gcr.io/runconduit/base:2017-10-30.01
     ---> 14aa74f25501
    Step 2/6 : RUN apt-get update
     ---> Running in c3b174baabea
    Get:1 http://security.debian.org jessie/updates InRelease [63.1 kB]
    Ign http://deb.debian.org jessie InRelease
    Get:2 http://deb.debian.org jessie-updates InRelease [145 kB]
    Get:3 http://deb.debian.org jessie Release.gpg [2434 B]
    Get:4 http://deb.debian.org jessie Release [148 kB]
    Get:5 http://security.debian.org jessie/updates/main amd64 Packages [640 kB]
    Get:6 http://deb.debian.org jessie-updates/main amd64 Packages [23.1 kB]
    Get:7 http://deb.debian.org jessie/main amd64 Packages [9064 kB]
    Fetched 10.1 MB in 1min 33s (107 kB/s)
    Reading package lists...
     ---> 52a8e6b98213
    Removing intermediate container c3b174baabea
    Step 3/6 : RUN apt-get install -y ca-certificates
     ---> Running in d0c086411dcb
    Reading package lists...
    Building dependency tree...
    Reading state information...
    The following extra packages will be installed:
      openssl
    The following NEW packages will be installed:
      ca-certificates openssl
    0 upgraded, 2 newly installed, 0 to remove and 7 not upgraded.
    Need to get 872 kB of archives.
    After this operation, 1495 kB of additional disk space will be used.
    Get:1 http://deb.debian.org/debian/ jessie/main openssl amd64 1.0.1t-1+deb8u7 [665 kB]
    Get:2 http://deb.debian.org/debian/ jessie/main ca-certificates all 20141019+deb8u3 [207 kB]
    debconf: delaying package configuration, since apt-utils is not installed
    Fetched 872 kB in 2s (399 kB/s)
    Selecting previously unselected package openssl.
    (Reading database ... 7934 files and directories currently installed.)
    Preparing to unpack .../openssl_1.0.1t-1+deb8u7_amd64.deb ...
    Unpacking openssl (1.0.1t-1+deb8u7) ...
    Selecting previously unselected package ca-certificates.
    Preparing to unpack .../ca-certificates_20141019+deb8u3_all.deb ...
    Unpacking ca-certificates (20141019+deb8u3) ...
    Setting up openssl (1.0.1t-1+deb8u7) ...
    Setting up ca-certificates (20141019+deb8u3) ...
    debconf: unable to initialize frontend: Dialog
    debconf: (TERM is not set, so the dialog frontend is not usable.)
    debconf: falling back to frontend: Readline
    debconf: unable to initialize frontend: Readline
    debconf: (Can't locate Term/ReadLine.pm in @INC (you may need to install the Term::ReadLine module) (@INC contains: /etc/perl /usr/local/lib/x86_64-linux-gnu/perl/5.20.2 /usr/local/share/perl/5.20.2 /usr/lib/x86_64-linux-gnu/perl5/5.20 /usr/share/perl5 /usr/lib/x86_64-linux-gnu/perl/5.20 /usr/share/perl/5.20 /usr/local/lib/site_perl .) at /usr/share/perl5/Debconf/FrontEnd/Readline.pm line 7.)
    debconf: falling back to frontend: Teletype
    Updating certificates in /etc/ssl/certs... 174 added, 0 removed; done.
    Processing triggers for ca-certificates (20141019+deb8u3) ...
    Updating certificates in /etc/ssl/certs... 0 added, 0 removed; done.
    Running hooks in /etc/ca-certificates/update.d....done.
     ---> ac2e44a7a879
    Removing intermediate container d0c086411dcb
    Step 4/6 : RUN mkdir /app
     ---> Running in ca3caf62eb9e
     ---> 73a3a5e1fb09
    Removing intermediate container ca3caf62eb9e
    Step 5/6 : ADD target/bb /app/
     ---> 55e2defdcf60
    Step 6/6 : ENTRYPOINT /app/bb
     ---> Running in f4f571b01dd8
     ---> e6d76c5df612
    Removing intermediate container f4f571b01dd8
    Successfully built e6d76c5df612
    Successfully tagged gcr.io/runconduit/bb:v1

A test run using the Docker CLI should return usage information and confirm everything is ok:

     $ docker run gcr.io/runconduit/bb:v1
    Various microservices that can be used to build a test lab for Conduit
    
    Usage:
      bb [command]
    
    Available Commands:
      broadcast-channel      Forwards the request to all downstream services.
      help                   Help about any command
      http-egress            Receives a request, makes a HTTP(S) call to a specified URL and return the body of the response
      point-to-point-channel Forwards the request to one and only one downstream service.
      terminus               Receives the request and returns a pre-defined response
    
    Flags:
          --downstream-timeout duration          timeout to use when making downstream connections. (default 1m0s)
          --fire-and-forget                      do not wait for a response when contacting downstream services.
          --grpc-downstream-server stringSlice   list of servers (hostname:port) to send messages to using gRPC, can be repeated
          --grpc-server-port int                 port to bind a gRPC server to (default -1)
          --h1-downstream-server stringSlice     list of servers (protocol://hostname:port) to send messages to using HTTP 1.1, can be repeated
          --h1-server-port int                   port to bind a HTTP 1.1 server to (default -1)
      -h, --help                                 help for bb
          --id string                            identifier for this container
          --log-level string                     log level, must be one of: panic, fatal, error, warn, info, debug (default "debug")
          --percent-failure int                  percentage of requests that this service will automatically fail
          --sleep-in-millis int                  amount of milliseconds to wait before actually start processing as request
    
    Use "bb [command] --help" for more information about a command.

To build the exact same scenario we had above, but for Kubernetes, you should have a YAML 
configuration like the following:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: bb-readme
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: bb-readme-terminus
  namespace: bb-readme
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bb-readme-terminus
  template:
    metadata:
      labels:
        app: bb-readme-terminus
    spec:
      containers:
      - name: http-to-grpc
        image: gcr.io/runconduit/bb:v1
        args: ["terminus", "--grpc-server-port", "9090", "--response-text", "BANANA"]
        ports:
        - containerPort: 9090
---
apiVersion: v1
kind: Service
metadata:
  name: bb-readme-terminus-svc
  namespace: bb-readme
spec:
  selector:
    app: bb-readme-terminus
  ports:
  - name: grpc
    port: 9090
    targetPort: 9090
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: bb-readme-gateway
  namespace: bb-readme
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bb-readme-gateway
  template:
    metadata:
      labels:
        app: bb-readme-gateway
    spec:
      containers:
      - name: http-to-grpc
        image: gcr.io/runconduit/bb:v1
        args: ["point-to-point-channel", "--grpc-downstream-server", "bb-readme-terminus-svc:9090", "--h1-server-port", "8080"]
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: bb-readme-gateway-svc
  namespace: bb-readme
spec:
  selector:
    app: bb-readme-gateway
  type: LoadBalancer
  ports:
  - name: http
    port: 8080
    targetPort: 8080
```

If you copy and paste the YAML content above in a file called `application.yaml`, you can 
then deploy it to your Kubernetes cluster by running:

    $ kubectl apply -f application.yaml

You can then use `curl`to query the service:

    $ curl `minikube -n bb-readme service bb-readme-gateway-svc  --url`
    {"requestUid":"in:http-sid:point-to-point-channel-grpc:-1-h1:8080-66349706","payload":"BANANA"}
