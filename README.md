# Goomerang

A [protocol buffers](https://developers.google.com/protocol-buffers/) over [websocket](https://datatracker.ietf.org/doc/html/rfc6455)
communications library.
<br>
<br>
<p align="center">
<img src="https://raw.githubusercontent.com/lidiackr/gophers/main/eloylp/goomerang/head.png" alt="goomerang" width="550"/>
</p>

<p align="right" style="color:silver">
Gopher art by <a href="https://github.com/lidiackr">@lidiackr</a>
</p>

## Status

This project is still a proof of concept. It has none or very little production experience. Maintainers of the project reserve the right of
breaking the public API.

## Motivation

Goomerang is an effort to provide an easy, application friendly way for communicating clients and servers. Because some of us only need to
send some protos over a fast pipe.

![idea](docs/idea.drawio.svg)

Possible status transitions for servers:

```mermaid
graph LR;
NEW-->RUNNING-->CLOSING-->CLOSED;
```

Possible status transitions for clients:

```mermaid
graph LR;
NEW-->RUNNING-->CLOSING-->CLOSED;
CLOSED-->RUNNING;
```

## Main features

* Simple, modular, embeddable. Use only the parts of interest. Have full control over the dependency.
* Websockets as underlying transport (currently built on top of [gorilla/websocket](https://github.com/gorilla/websocket)).
* [Protocol buffers](https://developers.google.com/protocol-buffers/) as first class citizen. Register [message handlers](#message-handlers)
  at clients and
  servers.
* [Middleware support](#middlewares), based on the Go HTTP standard lib.
* [Broadcast messages](#broadcasts) from the server to clients.
* Send [synchronous messages](#synchronous-sends) from the client side, following a request/response pattern.
* Support for concurrency at a message handler level.
* Support for configurable [hooks](#hooks) for certain actions.
* [Graceful shutdown](#graceful-shutdown) of the websocket connection/processing.
* [Ping/pong](https://datatracker.ietf.org/doc/html/rfc6455#section-5.5.2) messages out of the box in clients and server. Keep alive
  connections.
* Customizable TLS configuration.
* Custom errors are exposed, ability to build retry systems on top of the public API.
* [Observability tools](#observability), that will help to instrument clients and servers.

## Installation

```bash
go get go.eloylp.dev/goomerang
```

## Basic usage

If you are unfamiliar with protocol buffers, all the examples on this readme plus a little guide can be found [here](example/protos/README.md).

Let's create a basic server. We just need to import the server package from the library. The server has many options
available, we encourage the user to review them [here](server/opts.go).

```go
package main

import (
	"log"

	"go.eloylp.dev/goomerang/server"
)

func main() {
	s, err := server.New(
		server.WithListenAddr("127.0.0.1:8080"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Run(); err != nil { // Will block the program till exit.
		log.Fatal(err)
	}
	//...
}
```

The client side package its quite symmetric with the server one. The client has also many configurable options, see
them [here](client/opts.go). Let's see a client example:

```go
package main

import (
	"context"
	"log"

	"go.eloylp.dev/goomerang/client"
)

func main() {
	//...
	c, err := client.New(
		client.WithServerAddr("127.0.0.1:8080"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	//...
}
```

That's it ! we just have connected a client and a server ! But not in a very
useful way. Both, client and server have support for handlers. See the next sections for more information.

## Messages

Goomerang facilitates the creation of custom protocols in messages. Messages can have 2 parts, `headers` and a `payload`. Headers are just
a key/value map of type string/string. The payload allows holding any type which accomplishes
the [proto.Message](https://github.com/protocolbuffers/protobuf-go/blob/v1.28.0/proto/proto.go#L24) interface, so any protocol buffer
message type.

```go
package main

import (
	"go.eloylp.dev/goomerang/message"

	"go.eloylp.dev/goomerang/example/protos"
)

func main() {
	msg := message.New().
		SetHeader("status", "requested").
		SetPayload(&protos.MessageV1{})
}
```

It's encouraged to use the message builder, so all complexities and needed data will be fulfilled.

The `message.Message` type it's normally used in message handlers. It can also be used
in the public API of client and server instances for simple operations like `Send(msg)`. See more in the following documentation.

## A history of handlers and middlewares

There's support for handlers and middlewares for both, client and server sides. This part of the project comes inspired by the
Go [HTTP standard library](https://pkg.go.dev/net/http#HandlerFunc). The intention is making this part familiar and versatile.

Handlers and middlewares **can only be registered before starting** the client or the server.

### Message Handlers

Let's see how to register a message handler.

```go
package main

import (
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
	"go.eloylp.dev/goomerang/example/protos"
)

func main() {

	// Create the server
	s, _ := server.New(server.WithListenAddr("127.0.0.1:8080"))

	// protos.MessageV1 represents a hypothetical message in your proto repo.
	s.Handle(&protos.MessageV1{}, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		// We need to cast the message, from the proto interface, to the concrete message type.
		msgT := msg.Payload.(*protos.MessageV1)

		// Do whatever with your message,
		// your handling logic goes here.

		// Alternatively, reply with another message.
		payload := &protos.ReplyV1{}
		reply := message.New().
			SetPayload(payload).
			SetHeader("status", "200")

		_, err := sender.Send(reply) // Replies to the client connection.
		// check for errors ...
	}))
}
```

Symmetrically, the same handler registering signature can be found at the client public API.

As handlers only need to accomplish the `message.Handler` interface, It's very common to use structs for holding handler dependencies,
which need to be thread safe:

```go
package main

import (
	"database/sql"

	"go.eloylp.dev/goomerang/message"
)

type Handler struct {
	DB *sql.DB
}

func (m *Handler) Handle(sender message.Sender, msg *message.Message) {
	//  handling logic goes there.
}
```

### Middlewares

Middlewares are just message handlers that always get executed no matter the kind of message, providing the ability to execute the next
handler in the chain. They are
being executed just before the message handler, bringing the opportunity to add extra logic to the message processing like metrics, logging
or panic handlers. Let's see how to register a middleware:

```go
package main

import (
	"fmt"

	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func main() {
	s, _ := server.New(server.WithListenAddr("127.0.0.1:8080"))
	s.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
			fmt.Printf("received message of kind: %q", msg.Metadata.Kind)
			h.Handle(sender, msg) // Continue with the next handler in chain.
		})
	})
}
```

That's sounds really familiar ! Yes, Goomerang tries to preserve same aspects of the standard library. Again, the same symmetric interface
can be found in the client public API.

It's important to note that **any number of middlewares can be registered**. The **order of registration will drive the order of execution**
. So for example, if we had:

```go
package main

import (
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func main() {
	s, _ := server.New(server.WithListenAddr("127.0.0.1:8080"))
	s.Handle(Handler())
	s.Middleware(Middleware1())
	s.Middleware(Middleware2())
}
```

The order of execution would be:

```mermaid
graph LR;
Middleware1-->Middleware2;
Middleware2-->Handler;
```

This library facilitates some middleware implementations [here](./middleware) that can be used directly. Let's see the example of the panic
handler one:

```go
package main

import (
	"fmt"

	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/middleware"
	"go.eloylp.dev/goomerang/server"
)

func main() {
	s, _ := server.New(server.WithListenAddr("127.0.0.1:8080"))
	s.Middleware(middleware.Panic(func(p interface{}) {
		fmt.Printf("panic detected: %v", p)
	}))
}
```

As a rule of thumb, its recommended this middleware to be implemented as first in the chain (so it protects the rest of handlers), in
order to not crash the entire process if
a panic arises at some point in the handler chain.

## Broadcasts

From the server side perspective, it's possible to send a message to all connected clients. This can be useful for a number of cases, like
push notifications. Here's and example on how to do so:

```go
package main

import (
	"context"
	"log"

	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
	"go.eloylp.dev/goomerang/example/protos"
)

func main() {
	// Create the server
	s, _ := server.New(server.WithListenAddr("127.0.0.1:8080"))

	msg := message.New().SetPayload(&protos.MessageV1{})

	_, err := s.BroadCast(context.TODO(), msg)
	if err != nil {
		log.Fatal(err)
	}
}
```

From the client side, it is not possible to send broadcasts to all the other connected clients. However, nothing will stop the
developer to create a handler for enabling this operation in the server. An example can be found [here](client_side_broadcast_test.go).

## Synchronous sends

The default send methods `c.Send()` `s.Broadcast()` from client and server respectively, **are completely asynchronous**. They work as a
"fire and forget" send system. That means the user of the library should design a way to ensure the message was received and processed by
the server before removing it from its internal state. Unless of course, the intrinsic value of the data expires very quick,
and the client can afford its lost.

In response to this issue, clients on this library have a `c.SyncSend(ctx, msg)` method implemented. This allows clients sending messages
and wait for the server
reply.

Here is an example of its use:

```go
package main

import (
	"context"
	"log"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/example/protos"
)

func main() {
	c, err := client.New(client.WithServerAddr("127.0.0.1:8080"))
	if err != nil {
		log.Fatal(err)
	}

	msg := message.New().SetPayload(&protos.MessageV1{})
	// Will block till reply from the server is received or
	// the context is cancelled.
	_, respProto, err := c.SendSync(context.TODO(), msg)
	if err != nil {
		log.Fatal(err)
	}

	if respProto.GetHeader("status") != "OK" {
		resp := respProto.Payload.(*protos.BadReplyV1)
		// Do something with the bad reply
		return
	}
	resp := respProto.Payload.(*protos.SuccessReplyV1)
	// Do something with the success reply
}
```

This is just a high level request/response pattern built on the top of the "fire and forget"
previously commented send system. It probably can help some clients, when the reply is
crucial for completing the operation.

Asynchronous methods just write to the TCP socket send buffer. They are only going to block the call if the other peer stops ACKing
TCP packets, so the pre-negotiated TCP window size is exceeded. That way the sender knows when to stop sending messages to the other peer,
until it starts performing more TCP acks again.

The following command can give us minimum, default and maximum receive buffers size in bytes in a Linux machine.

```bash
$ cat /proc/sys/net/ipv4/tcp_rmem 
4096	131072	6291456
```

## Hooks

Sometimes getting feedback from internal parts of the system its difficult. Specially in processing loops, where in case of errors we cannot
return them to the user, and we do not want to make decisions on their place. To deal with this,
in both parts (client and server) the user can register function hooks. A complete list of them can be found in the configuration options of
both, [client](client/opts.go) and [server](server/opts.go). Let's take a look on how to register an error hook:

```go
package main

import (
	"log"

	"go.eloylp.dev/goomerang/server"
)

func main() {
	s, _ := server.New(
		server.WithListenAddr("127.0.0.1:8080"),
		server.WithOnErrorHook(func(err error) {
			log.Printf("logging error: %v", err)
		}),
		server.WithOnErrorHook(func(err error) {
			// Do a second action with err.
		}),
	)
}
```

This error hooks are very handy to log errors, so the user can use custom loggers/metrics registries. All the hooks can be
registered multiple times. The logic behind will just execute the hooks in order.

## Graceful shutdown

The current implementation supports a graceful shutdown
as described
in [RFC6455](https://datatracker.ietf.org/doc/html/rfc6455#section-1.4). One side of the connections initiates the shutdown procedure,
signaling it's not going to send more messages and waiting for the same reply from the other peer. Once all processing logic (like handlers)
end, the library just returns the control to the user. This implementation works in a best-effort way, many things can go wrong in the
middle of the process. Users of the library are encouraged to not rely on this shutdown procedure regarding data integrity.

Let's see the shutdown process when a **client initiates** the shutdown procedure:

```mermaid
sequenceDiagram
participant C1 as Client-1
participant S as Server(c-1)
Note left of C1: Status = Running
Note left of C1: Status = Closing
C1->>S: Close signal
Note right of S: Status = Running
alt if server is responsive
S->>C1: Close signal
par
Note right of S: Status = Running
S->>S: Processing shutdown for c1
S->>C1: Connection close for c1
end
end
C1->>C1: Processing shutdown
alt in case server is unresponsive (no previous close signal received)
C1->>S: Connection close
end
Note left of C1: Status = Closed
Note right of S: Status = Running
C1->>C1: Returns to user
```

The server never changes its status, as the initiator of the shutdown was only one client. So only the "connection slot" and processing for
that client is removed from the server.

Let's check now the sequence when the **server initiates** the shutdown connection:

```mermaid
sequenceDiagram
participant S as Server
participant C1 as Client-1
participant C2 as Client-2
Note left of S: Status = Running
Note left of S: Status = Closing
par
Note right of C1: Status = Running
S->>C1: Close signal
Note right of C1: Status = Closing
Note right of C2: Status = Running
S->>C2: Close signal
Note right of C2: Status = Closing
end
alt clients are responsive (some of them could timeout)
par
C1->>S: Close signal
C1->>C1: Processing shutdown
Note right of C1: Status = Closed

C2->>S: Close signal
C2->>C2: Processing shutdown 
Note right of C2: Status = Closed

end
end
S->>S: Processing shutdown 
Note left of S: Status = Closed
S->>S: Return to user
```

## Observability

### Prometheus metrics

By default, this library assumes no metrics. However, it provides tools for a quick instrumentation with
[Prometheus](https://prometheus.io/) metrics for both, [clients](metrics/client.go) and [servers](metrics/server.go). Let's see an
example of a metered client:

```go
package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/metrics"
)

func main() {
	// Metrics creation, with the default config.
	m := metrics.NewClientMetrics(metrics.DefaultClientConfig())

	// Register the metrics in Prometheus registry. This time, the global one.
	m.Register(prometheus.DefaultRegisterer)

	// Use the same options as in a normal client constructor.
	c, err := client.NewMetered(m,
		client.WithServerAddr("127.0.0.1:9090"),
		client.WithMaxConcurrency(5),
		// ...
	)
	if err != nil {
		log.Fatal(err)
	}
	// ...
}
```

Note in the above example, the `client.NewMetered(...)` accepts as second argument the same type of [options](client/opts.go) exposed in the
[basic example](#basic-usage).

As commented in the [middleware section](#middlewares), the user also has access to all low level observability middlewares in
the [middleware package](middleware). We encourage the user to take a look there in case further customization its needed.

The same, symmetric interface can be found in the server public API.

### Grafana dashboard

![img.png](docs/grafana-dashboard.png)

This library provides an
out-of-the-box [Grafana](https://grafana.com/) [dashboard](internal/lab/grafana/provisioning/dashboards/goomerang.json) which should cover
the
basic usage. Users of the library are encouraged to adapt it at discretion.

# Contributing

Contributions are welcome ! If you think something could be improved, request a new feature or just want to leave some feedback,
please check our [contributing](CONTRIBUTING.md) guide. 