# gRPC OPA

This is a library providing an interceptor for gRPC servers to authorize client requests against an *O*pen *P*olicy *A*gent HTTP Server.

## How it works

The gRPC interceptors intercept every request and check for a certain header (`metadata`) to be present. If so the value is extracted and HTTP POSTed to the configured OPA backend route together with the invoked gRPC (full-)method.

Having a policy like this in place could then be used to authorize your gRPC APIs.

```rego
default invocation_allowed = false

invocation_allowed = true {
    clients[input.authToken][_] = input.method
}

clients = {
    "123": ["/helloworld.Greeter/SayHello"],
    "234": ["/helloworld.Greeter/SayHello", "/admins/Purge"],
    "456": [],
}
```

## Getting Started

Get the lib: `go get github.com/typusomega/grpc-opa`

Afterwards register an `OpaAuthorizer` in your gRPC server like this:

```go
import (
	// ...
    opaauthz "github.com/typusomega/grpc-opa/opa"
    // ...
)

// ...
authz := opaauthz.NewOpaAuthorizer(opaauthz.OpaURL("http://localhost:8181/v1/data/apis/invocation_allowed"))

s := grpc.NewServer(
    grpc.StreamInterceptor(authz.OpaStreamInterceptor),
    grpc.UnaryInterceptor(authz.OpaUnaryInterceptor),
)

pb.RegisterGreeterServer(s, &server{})
if err := s.Serve(lis); err != nil {
    log.Fatalf("failed to serve: %v", err)
}
```

There are complete examples for [server](./example-server/server.go) and [client](./example-client/client.go) in the respective directories.





