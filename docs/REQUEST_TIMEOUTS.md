# Request cancellation and response ownership

The Go API gives each request a 25-second cooperative deadline. An earlier parent
deadline wins, and a client disconnect cancels the same context. Handlers run on
the normal Gin request goroutine; the timeout middleware never starts a second
handler goroutine, replaces the response writer, or sends its own error response.

These are separate limits:

| Limit | Current setting | What it controls |
| --- | --- | --- |
| Request context | 25 seconds | Cancellation of context-aware application work |
| HTTP write timeout | 30 seconds | Network writes for the response |
| HTTP read timeout | 15 seconds | Reading the request, including its body |
| HTTP header timeout | 5 seconds | Reading request headers |
| HTTP idle timeout | 60 seconds | Waiting for the next keep-alive request |

The settings live in `go-b2b-starter/internal/platform/server/domain/middleware.go`
and `http_server.go`. The write deadline is measured by `net/http`, not from the
moment the application finishes its work; it is not an extra 30 seconds after the
request deadline. A write timeout can surface as a closed or truncated connection,
and small writes may first succeed into an HTTP buffer. It is not a reliable JSON
504 response or a way to interrupt arbitrary Go code.

## Adding a handler or integration

Pass `c.Request.Context()` through services to SQL transactions, queries and
outbound HTTP requests. Bound provider clients as well. Loops and streams must
check cancellation before starting their next unit of work and return when the
context ends. Do not replace the request context with `context.Background()` or
retain the Gin context for a detached goroutine. If request work starts a
goroutine, it must cancel and join that goroutine before returning.

The handler owns the response, including translating a canceled dependency into
an appropriate existing endpoint error. The middleware does not promise a uniform
504 response. Before any output has been committed, a handler can send an error;
after headers or a stream chunk have been written, stop the operation and return.
Do not append a second JSON error to a partial response or change its status.
Cancellation does not roll back an already committed mutation or undo an external
provider operation, so a timeout is not evidence that retrying a mutation is safe.

There is intentionally no hard completion guarantee for a handler that ignores
its context. Such a handler can continue running and attempting writes after the
deadline. Moving it into a goroutine and returning early would allow it to use a
response writer or recycled Gin context after the request has ended. A timeout
must not reintroduce that ownership bug.

Work that routinely exceeds the request budget, must survive a client disconnect,
or needs durable retries belongs in an explicitly designed background job with a
persisted status, tenant scope, idempotency and cancellation/retry policy. Add that
facility when a product feature requires it; the starter does not need a queue for
ordinary organization, member, profile or billing requests. An HTTP stream needs
its own deliberate lifetime and write-deadline design rather than disabling limits
globally.

## Verification

From `go-b2b-starter/`, run:

```sh
go test -race ./internal/platform/server/middleware
go test -race ./...
go vet ./...
```

The timeout tests use `testing/synctest` for deterministic request deadlines,
earlier parent deadlines, parent cancellation, successful-response cleanup,
partial and flushed responses, and the non-cooperative-handler limit. They also
exercise panic recovery, real HTTP client disconnect, and a server write deadline
that rejects a late socket write without interrupting the handler's work.
They establish bounded completion for the cooperative test handlers, without
claiming that all future integrations necessarily honor context cancellation.

References: [Go request contexts](https://pkg.go.dev/net/http#Request.Context),
[HTTP server timeouts](https://pkg.go.dev/net/http#Server), and
[deterministic concurrency tests](https://pkg.go.dev/testing/synctest).
