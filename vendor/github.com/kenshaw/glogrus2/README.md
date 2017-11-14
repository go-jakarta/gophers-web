# goji/glogrus [![GoDoc](https://godoc.org/github.com/goji/glogrus2?status.png)](https://godoc.org/github.com/goji/glogrus2)

glogrus2 provides structured logging via logrus for Goji2.

## Example

**Simple logging**
```go

package main

import(
	"github.com/goji/glogrus2"
    "goji.io"
    "net/http"
    "github.com/sirupsen/logrus"
)

func main() {
    router := goji.NewMux()
	logr := logrus.New()
	logr.Formatter = new(logrus.JSONFormatter)
	goji.Use(glogrus.NewGlogrus(logr, "my-app-name"))

	log.Fatal(http.ListenAndServe(":8080", router))
}

```

**Logging with custom request id from http Context**
```go

package main

import(
	"github.com/goji/glogrus2"
    "goji.io"
    "golang.org/x/net/context"
    "net/http"
    "github.com/sirupsen/logrus"
)

func main() {
    router := goji.NewMux()
	logr := logrus.New()
	logr.Formatter = new(logrus.JSONFormatter)
	router.UseC(glogrus.NewGlogrusWithReqId(logr, "my-app-name", IdFromContext))

	log.Fatal(http.ListenAndServe(":8080", router))
}

func IdFromContext(ctx context.Context) string {
    return ctx.Value("requestIdKey")
}
```
- - -
#### Need something to put requestId in your Context?
[gojiid can help you with that](https://github.com/atlassian/gojiid)

#### Looking for hierarchical structured logging?
[slog](https://github.com/zenazn/slog) and [lunk](https://github.com/codahale/lunk) looks interesting.
