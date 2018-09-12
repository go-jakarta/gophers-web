package logrusmw

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// NewWithID creates standard HTTP middleware that logs all requests and
// responses using the supplied logger.
func NewWithID(logger *logrus.Logger, idFunc func(context.Context) string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start, ctxt := time.Now(), req.Context()
			id := idFunc(ctxt)

			logger.WithFields(logrus.Fields{
				"req_id": id,
				"uri":    req.RequestURI,
				"method": req.Method,
				"remote": req.RemoteAddr,
			}).Info("req_start")

			rw := &resWriter{ResponseWriter: res}
			h.ServeHTTP(rw, req)
			rw.maybeWriteHeader()

			latency := float64(time.Since(start)) / float64(time.Millisecond)

			logger.WithFields(logrus.Fields{
				"req_id":  id,
				"uri":     req.RequestURI,
				"method":  req.Method,
				"remote":  req.RemoteAddr,
				"status":  rw.status,
				"latency": fmt.Sprintf("%6.4f ms", latency),
			}).Info("req_served")
		})
	}
}

// New creates standard HTTP middleware that logs all requests and
// responses using the supplied logger.
func New(logger *logrus.Logger) func(http.Handler) http.Handler {
	return NewWithID(logger, func(context.Context) string { return "" })
}

// resWriter wraps a standard http.ResponseWriter with a flushed flag and the
// status result.
type resWriter struct {
	http.ResponseWriter
	flushed bool
	status  int
}

// WriteHeader stores the status and marks the writer as flushed.
func (rw *resWriter) WriteHeader(status int) {
	if !rw.flushed {
		rw.flushed, rw.status = true, status
		rw.ResponseWriter.WriteHeader(status)
	}
}

// Write writes the bytes and calls MaybeWriteHeader.
func (rw *resWriter) Write(buf []byte) (int, error) {
	rw.maybeWriteHeader()
	return rw.ResponseWriter.Write(buf)
}

// maybeWriteHeader writes the header if it is not alredy set.
func (rw *resWriter) maybeWriteHeader() {
	if !rw.flushed {
		rw.WriteHeader(http.StatusOK)
	}
}
