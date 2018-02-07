# building_blocks

Building Blocks or `bb` is a container that runs a server that can be configured 
to simulate many of the typical scenarios in a Service-Oriented Architecture.

## Using `bb`
`bb` publishes a single container, `buoyantio/bb:v1`. The containers receive and 
return a simple message, described by the 
[protobuf schema in this repository](api.proto). This known interface allows 
for `bb` containers to be arranged in many different ways, like LEGO blocks.

The best way to find out about `bb` features is by using the `--help` command 
and checking the self-contained documentation.

### Running locally
If you want to run `bb` locally, you need to build a binary for your 
environment. The usual way to go about this is using Go's standard build tools.
From the root of this repository, run: 

    $ mkdir -p target && go build -o target/bb ./building_blocks
    
This will create a `bb` binary in the `target` directory. You can run this on 
your computer but, unless you happen to use Linux, you won't be able to use 
this same binary on Docker or Kubernetes. 

As an example, to create a simple two-service setup open a terminal and type 
in the following:

    $  target/bb terminus --grpc-server-port 9090 --response-text BANANA
    
The on a second terminal type this:
    
    $ target/bb point-to-point-channel --grpc-downstream-server localhost:9090 --h1-server-port 8080
    
Now, on a third terminal, type this and you should see a similar response:
    
    $ curl localhost:8080
    {"requestUid":"in:http-sid:point-to-point-channel-grpc:-1-h1:8080-387107000","payload":"BANANA"}
    
The first command created a gRPC server on port 9090 following the 
[terminus strategy](strategies/terminus.go), which will return the payload 
defined as the `--response-text` argument for any valid request.

The second command creates a HTTP 1.1 server on port 8080 following the 
[point-to-point channel strategy](strategies/point_to_point_channel.go), and 
has the gRPC server previously defined as its downstream. It will receive any 
HTTP 1.1 request, convert it to gRPC, and forward it to the downstream server.
It will also get the gRPC response from the server, convert it to JSON over 
HTTP, and return to its client.
    
## Running on Kubernetes
Although `bb` can be useful to test things locally as described above, its main 
use case is to create complicated environments inside Kubernetes clusters.

To use `bb` with Kubernetes, the first step you need to take is to publish its 
Docker image to your Kubernetes cluster. Here we will be using a local Minikube 
installation to demonstrate its use.

First, make sure that Minikube is running:

    $ minikube status
      minikube: Running
      cluster: Running
      kubectl: Correctly Configured: pointing to minikube-vm at 192.168.99.100
      
Now, make sure that your Docker environment variables are set to use Minikube 
as the Docker repository for images:

    $ eval "$(minikube docker-env)"          
    
You should then build a Docker image for `bb`:

    $ bin/docker-build-building-blocks.sh
      Sending build context to Docker daemon  120.1MB
      Step 1/4 : FROM golang:latest
       ---> 3858fd70eed2
      Step 2/4 : RUN mkdir /app
       ---> Using cache
       ---> fb04b80efe54
      Step 3/4 : ADD target/bb /app/
       ---> d345414ca55c
      Removing intermediate container 03c6268458b2
      Step 4/4 : CMD /app/bb
       ---> Running in 1b10a0b340b2
       ---> cac5ea6b1538
      Removing intermediate container 1b10a0b340b2
      Successfully built cac5ea6b1538
      Successfully tagged buoyantio/bb:v1   
      
A test run using the Docker CLI should return usage information and confirm 
everything is ok:

    $ docker run buoyantio/bb:v1
    Various microservices that can be used to build a test lab for Conduit
    
    Usage:
      bb [command]
    
    Available Commands:
      broadcast-channel      Forwards the request to all downstream services.
      help                   Help about any command
      point-to-point-channel Forwards the request to one and only one downstream service.
      terminus               Receives the request and returns a response
    
    Flags:
          --grpc-downstream-server stringSlice   list of servers (hostname:port) to send messages to using gRPC, can be repeated
          --grpc-server-port int                 port to bind a gRPC server to (default -1)
          --h1-downstream-servers stringSlice    list of servers (hostname:port) to send messages to using HTTP 1.1, can be repeated
          --h1-server-port int                   port to bind a HTTP 1.1 server to (default -1)
      -h, --help                                 help for building_blocks
          --id string                            identifier for this container
          --log-level string                     log level, must be one of: panic, fatal, error, warn, info, debug (default "debug")
          --percent-failure int                  percentage of requests that this service will automatically fail
          --sleep-in-millis int                  amount of milliseconds to wait before actually start processing as request
    
    Use "bb [command] --help" for more information about a command.      
    
To build the exact same scenario we had above, but for Kubernetes, you should have 
a YAML configuration like the following:

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
        image: buoyantio/bb:v1
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
        image: buoyantio/bb:v1
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

If you copy and paste the YAML content above in a file called 
`application.yaml`, you can then deploy it to ypour Kubernetes cluster by 
running:

    $ kubectl apply -f application.yaml

You can then use `curl`to query the service:

    $ curl `minikube -n bb-readme service bb-readme-gateway-svc  --url`
    {"requestUid":"in:http-sid:point-to-point-channel-grpc:-1-h1:8080-66349706","payload":"BANANA"}
      
