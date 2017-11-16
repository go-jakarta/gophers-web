// Package secure is an HTTP middleware for Go that handles adding security
// headers to HTTP responses, and accompanying security checks.
//
//    package main
//
//    import (
//        "net/http"
//
//        "github.com/kenshaw/secure"
//    )
//
//    var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//        w.Write([]byte("hello world"))
//    })
//
//    func main() {
//        secureMiddleware := secure.New(
//            secure.AllowedHosts("www.example.com", "sub.example.com"),
//            secure.SSLRedirect(true),
//        })
//
//        app := secureMiddleware.Handler(myHandler)
//        http.ListenAndServe("127.0.0.1:3000", app)
//    }
package secure

//go:generate go run gen-readme.go > README.md

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Middleware that sets basic security headers and provides simple security
// checks for http servers.
type Middleware struct {
	// AllowedHosts is a list of fully qualified domain names that are allowed.
	// When empty, allows any host.
	AllowedHosts []string

	// HostsProxyHeaders is a set of header keys that may hold a proxied
	// hostname value for the request.
	HostsProxyHeaders []string

	// If SSLRedirect is set to true, then only allow https requests.
	SSLRedirect bool

	// If SSLTemporaryRedirect is true, the a 302 will be used while
	// redirecting.
	SSLTemporaryRedirect bool

	// SSLHost is the host name that is used to redirect http requests to
	// https. If not set, indicates to use the same host.
	SSLHost string

	// SSLForwardedProxyHeaders is the set of header keys with associated
	// values that would indicate a valid https request. This is used when
	// proxying requests from behind another webserver (ie, nginx, apache,
	// etc).
	//
	//     &secure.Middleware{
	//         SSLForwardedProxyHeaders: map[string]string{
	//             "X-Forwarded-Proto": "https",
	//         },
	//     }
	//
	SSLForwardedProxyHeaders map[string]string

	// STSSeconds is the max-age of the Strict-Transport-Security header.
	// Header will not be included if STSSeconds = 0.
	STSSeconds int64

	// When STSIncludeSubdomains is true, `includeSubdomains` will be appended to
	// the Strict-Transport-Security header.
	STSIncludeSubdomains bool

	// When STSPreload is true, the `preload` flag will be appended to the
	// Strict-Transport-Security header.
	STSPreload bool

	// When ForceSTSHeader is true, the STS header will be added even when the
	// connection is HTTP.
	ForceSTSHeader bool

	// When FrameDeny is true, adds the X-Frame-Options header with the value
	// of `DENY`.
	FrameDeny bool

	// CustomFrameOptionsValue allows the X-Frame-Options header value to be
	// set with a custom value. Overrides the FrameDeny option.
	CustomFrameOptionsValue string

	// If ContentTypeNosniff is true, adds the X-Content-Type-Options header
	// with the value `nosniff`.
	ContentTypeNosniff bool

	// If BrowserXSSFilter is true, adds the X-XSS-Protection header with the
	// value `1; mode=block`.
	BrowserXSSFilter bool

	// CustomBrowserXSSValue allows the X-XSS-Protection header value to be set
	// with a custom value. This overrides the BrowserXSSFilter option.
	CustomBrowserXSSValue string

	// ContentSecurityPolicy allows the Content-Security-Policy header value to
	// be set with a custom value.
	ContentSecurityPolicy string

	// ReferrerPolicy configures which the browser referrer policy.
	ReferrerPolicy string

	// BadHostHandler is the bad host handler.
	BadHostHandler http.HandlerFunc

	// When DevEnvironment is true, disables the AllowedHosts, SSL, and STS
	// checks.
	//
	// This should be toggled only when testing / developing, and is necessary
	// when testing sites configured only for https from a http based
	// connection.
	//
	// If you would like your development environment to mimic production with
	// complete Host blocking, SSL redirects, and STS headers, leave this as
	// false.
	DevEnvironment bool
}

// New constructs a new secure Middleware instance with the supplied options.
func New(opts ...Option) *Middleware {
	s := &Middleware{
		BadHostHandler: DefaultBadHostHandler,
	}

	// apply opts
	for _, o := range opts {
		o(s)
	}

	return s
}

// Handler implements the http.HandlerFunc for integration with the standard
// net/http lib.
func (s *Middleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// process the request. if an error is returned, then bail (ie, request
		// should not continue down handler chain)
		if err := s.Process(w, r); err != nil {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// HandlerFuncWithNext is a special implementation for Negroni, but could be
// used elsewhere.
func (s *Middleware) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := s.Process(w, r)

	// If there was an error, do not call next.
	if err == nil && next != nil {
		next(w, r)
	}
}

// Process runs the actual checks and returns an error if the middleware chain
// should stop.
func (s *Middleware) Process(w http.ResponseWriter, r *http.Request) error {
	// resolve the host for the request, using proxy headers if present
	host := r.Host
	for _, header := range s.HostsProxyHeaders {
		if h := r.Header.Get(header); h != "" {
			host = h
			break
		}
	}

	// allowed hosts check
	if len(s.AllowedHosts) > 0 && !s.DevEnvironment {
		isGoodHost := false
		for _, allowedHost := range s.AllowedHosts {
			if strings.EqualFold(allowedHost, host) {
				isGoodHost = true
				break
			}
		}

		if !isGoodHost {
			if s.BadHostHandler != nil {
				s.BadHostHandler(w, r)
			} else {
				DefaultBadHostHandler(w, r)
			}
			return errors.New("bad host")
		}
	}

	// determine if we are on https
	isSSL := r.URL.Scheme == "https" || r.TLS != nil
	if !isSSL {
		for k, v := range s.SSLForwardedProxyHeaders {
			if r.Header.Get(k) == v {
				isSSL = true
				break
			}
		}
	}

	// ssl check
	if s.SSLRedirect && !isSSL && !s.DevEnvironment {
		url := r.URL
		url.Scheme = "https"
		url.Host = host

		if len(s.SSLHost) > 0 {
			url.Host = s.SSLHost
		}

		status := http.StatusMovedPermanently
		if s.SSLTemporaryRedirect {
			status = http.StatusTemporaryRedirect
		}

		http.Redirect(w, r, url.String(), status)
		return fmt.Errorf("https redirect")
	}

	// only add Strict-Transport-Security header when we know it's an ssl
	// connection
	//
	// see: https://tools.ietf.org/html/rfc6797#section-7.2
	if s.STSSeconds != 0 && (isSSL || s.ForceSTSHeader) && !s.DevEnvironment {
		stsSub := ""
		if s.STSIncludeSubdomains {
			stsSub = "; includeSubdomains"
		}

		if s.STSPreload {
			stsSub += "; preload"
		}

		w.Header().Add("Strict-Transport-Security", fmt.Sprintf("max-age=%d%s", s.STSSeconds, stsSub))
	}

	// frame options
	if len(s.CustomFrameOptionsValue) > 0 {
		w.Header().Add("X-Frame-Options", s.CustomFrameOptionsValue)
	} else if s.FrameDeny {
		w.Header().Add("X-Frame-Options", "DENY")
	}

	// content type options
	if s.ContentTypeNosniff {
		w.Header().Add("X-Content-Type-Options", "nosniff")
	}

	// xss protection
	if len(s.CustomBrowserXSSValue) > 0 {
		w.Header().Add("X-XSS-Protection", s.CustomBrowserXSSValue)
	} else if s.BrowserXSSFilter {
		w.Header().Add("X-XSS-Protection", "1; mode=block")
	}

	// content security policy
	if len(s.ContentSecurityPolicy) > 0 {
		w.Header().Add("Content-Security-Policy", s.ContentSecurityPolicy)
	}

	// referrer policy
	if len(s.ReferrerPolicy) > 0 {
		w.Header().Add("Referrer-Policy", s.ReferrerPolicy)
	}

	return nil
}

// DefaultBadHostHandler is the default bad host http handler.
func DefaultBadHostHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "bad host", http.StatusInternalServerError)
}
