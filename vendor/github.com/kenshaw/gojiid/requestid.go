package gojiid

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

// not exported to avoid collision with other context keys
type key int

// requestIdKey is the context key for the request id.
const requestIdKey key = 0

// Type for configuring the middleware to get the request id using custom HTTP headers and/or a custom generator
// Headers is a slice of HTTP Header keys used to look for the request id in the HTTP request
// Generator is a custom function that will return a new request id if one is not already found in the HTTP headers
type RequestIdConfig struct {
	Headers   []string
	Generator func(req *http.Request) string
}

var prefix string
var idcounter uint64

// initializes default values for use with the default generator
func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// creates the default middleware that just generates an id using the default generator
func NewRequestId() func(http.Handler) http.Handler {
	config := &RequestIdConfig{make([]string, 0), defaultGenerator}
	return NewCustomRequestId(config)
}

// creates the middleware configured with custom Headers and Generator function
func NewCustomRequestId(config *RequestIdConfig) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqId, ok := fromHeaders(config.Headers, r)
			if !ok {
				reqId = config.Generator(r)
			}

			if len(reqId) > 0 {
				r = r.WithContext(context.WithValue(r.Context(), requestIdKey, reqId))
			}

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// Returns the request id that may have been previously set in the Context.
// If no request id was set, returns blank
func FromContext(ctx context.Context) string {
	requestId, ok := ctx.Value(requestIdKey).(string)

	if ok {
		return requestId
	} else {
		return ""
	}
}

func defaultGenerator(req *http.Request) string {
	myid := atomic.AddUint64(&idcounter, 1)
	return fmt.Sprintf("%s-%06d", prefix, myid)
}

func fromHeaders(headerKeys []string, req *http.Request) (string, bool) {
	var reqId string
	ok := false

	for _, key := range headerKeys {
		reqId = req.Header.Get(key)
		if len(reqId) > 0 {
			ok = true
			break
		}
	}
	return reqId, ok
}
