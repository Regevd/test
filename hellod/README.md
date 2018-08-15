# `hellod`

Cloud native distributed hellos.

### why

Why not?

### running locally

```
# build all service images locally
$ make build-images

# run dependencies (zookeeper, kafka, jaeger)
$ make deps

# run the services (in another terminal)
$ make run

# hit the endpoints
$ curl localhost:8080/hi/bob
{"message":"hi bob"}
$ curl localhost:8080/bye/bob
{"message":"bye bob"}
$ curl localhost:8080/hi/666
failed to call method: rpc error: code = InvalidArgument desc = oh no, it's the trap card

# view jaeger traces at http://localhost:16686/
# view raw prometheus metrics at http://localhost:8081/metrics (or any of the other service ports)

# cleanup
$ make clean
```

### what's happening?

- user hits `dude:8080/hi/bob` via curl/browser (HTTP)
  - `dude` hits `helloer:5050` via gRPC for a `hi` message
  - `dude` also hits `echoer:6060` via gRPC for a simple echo of the request
    - `echoer` responds with the input message, OR
    - `echoer` fails when receiving `666`, returning codes.InvalidArgument (gRPC's 400)
      - `echoer` writes the failed message to a failure topic in Kafka
  - `watcher` listens to the failure topic in Kafka, and logs incoming messages to stderr

- all services expose prometheus metrics
  - `dude` exposes HTTP server metrics and gRPC client metrics
  - `helloer` exposes gRPC server metrics
  - `echoer` exposes gRPC server metrics
  - `watcher` exposes custom domain metrics (failures observed)

- all messages are traced using OpenCensus, and collected by jaeger.  
  traces are propagated through all hops (http -> grpc -> kafka).
