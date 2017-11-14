package glogrus2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// New allows you to configure a goji middleware that logs all requests and
// responses using the structured logger logrus. It takes the logrus instance
// and the name of the app as the parameters and returns a middleware of type
// "func(http.Handler) http.Handler"
//
// Example:
//
//		package main
//
//		import(
//			""goji.io"
//			"github.com/goji/glogrus2"
//			"github.com/sirupsen/logrus"
//		)
//
//		func main() {
//
//			logr := logrus.New()
//			logr.Formatter = new(logrus.JSONFormatter)
//			goji.Use(glogrus.NewGlogrus(logr, "my-app-name"))
//
//			goji.Get("/ping", yourHandler)
//			goji.Serve()
//		}
//
func New(l *logrus.Logger, name string) func(http.Handler) http.Handler {
	return NewWithReqId(l, name, emptyRequestId)
}

// NewWithReqId allows you to configure a goji middleware that logs all
// requests and responses using the structured logger logrus. It takes the
// logrus instance, the name of the app and a function that can retrieve a
// requestId from the Context "func(context.Context) string" as the parameters
// and returns a middleware of type "func(http.Handler) http.Handler"
//
// Passing in the function that returns a requestId allows you to "plug in"
// other middleware that may set the request id.
//
// Example:
//
//		package main
//
//		import(
//			""goji.io"
//			"github.com/goji/glogrus2"
//			"github.com/sirupsen/logrus"
//		)
//
//		func main() {
//			logr := logrus.New()
//			logr.Formatter = new(logrus.JSONFormatter)
//			goji.Use(glogrus.NewWithReqId(logr, "my-app-name", GetRequestId))
//
//			goji.Get("/ping", yourHandler)
//			goji.Serve()
//		}
//
//		func GetRequestId(ctx context.Context) string {
//			return ctx.Value("requestIdKey")
//		}
//
func NewWithReqId(l *logrus.Logger, name string, reqidf func(context.Context) string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start, ctx := time.Now(), r.Context()

			reqID := reqidf(ctx)

			l.WithFields(logrus.Fields{
				"req_id": reqID,
				"uri":    r.RequestURI,
				"method": r.Method,
				"remote": r.RemoteAddr,
			}).Info("req_start")
			lresp := wrapWriter(w)

			h.ServeHTTP(lresp, r)
			lresp.maybeWriteHeader()

			latency := float64(time.Since(start)) / float64(time.Millisecond)

			l.WithFields(logrus.Fields{
				"req_id":  reqID,
				"status":  lresp.status(),
				"method":  r.Method,
				"uri":     r.RequestURI,
				"remote":  r.RemoteAddr,
				"latency": fmt.Sprintf("%6.4f ms", latency),
				"app":     name,
			}).Info("req_served")
		}
		return http.HandlerFunc(fn)
	}

}

func emptyRequestId(ctx context.Context) string {
	return ""
}
