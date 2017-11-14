# gojjid [![GoDoc](https://godoc.org/github.com/atlassian/gojiid?status.png)](https://godoc.org/github.com/atlassian/gojiid)

gojiid provides a customizable middleware to provide and retrieve a request id string via the http Context 

## Example

**Using the default generator**
```go

package main

import(
	"github.com/atlassian/gojiid"
	"github.com/goji/glogrus"
	"goji.io"
	"golang.org/x/net/context"
	"net/http"
)

func main() {
	router := goji.NewMux()
    
    //default request id generation
    router.UseC(gojiid.NewRequestId())
	
	logr := logrus.New()
	logr.Formatter = new(logrus.JSONFormatter)
	
	// add request id support to glogrus
	goji.UseC(glogrus.NewGlogrusWithReqId(logr, "myapp", gojiid.FromContext))

	log.Fatal(http.ListenAndServe(":8080", router))
}

```
**Grabbing the id from an http header**
```go

package main

import(
	"github.com/atlassian/gojiid"
	"github.com/goji/glogrus"
	"goji.io"
	"golang.org/x/net/context"
	"net/http"
)

func main() {
	router := goji.NewMux()
    
    // lookup from http headers in order
    router.UseC(gojiid.NewCustomRequestId(
        &gojiid.RequestIdConfig{Headers:[]string{"X-Request-ID", "my-custom-id-header"}}
    ))
	
	logr := logrus.New()
	logr.Formatter = new(logrus.JSONFormatter)
	
	// add request id support to glogrus
	goji.Use(glogrus.NewGlogrusWithReqId(logr, "myapp", gojiid.FromContext))

	log.Fatal(http.ListenAndServe(":8080", router))
}

```
**Grabbing the id from an http header and using a custom generator**
```go

package main

import(
	"github.com/atlassian/gojiid"
	"github.com/goji/glogrus"
	"goji.io"
	"golang.org/x/net/context"
	"net/http"
)

func main() {
	router := goji.NewMux()
    
    // lookup from http headers in order
    router.UseC(gojiid.NewCustomRequestId(
        &gojiid.RequestIdConfig{
            Headers:[]string{"X-Request-ID", "my-custom-id-header"},
            Generator: MyGenerator
        }
    ))
	
	logr := logrus.New()
	logr.Formatter = new(logrus.JSONFormatter)
	
	// add request id support to glogrus
	goji.Use(glogrus.NewGlogrusWithReqId(logr, "myapp", gojiid.FromContext))

	log.Fatal(http.ListenAndServe(":8080", router))
	
}

func MyGenerator(req *http.Request) string {
    return "static-id"
}
```