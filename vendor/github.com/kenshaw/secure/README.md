# About secure [![GoDoc](https://godoc.org/github.com/kenshaw/secure?status.svg)](http://godoc.org/github.com/kenshaw/secure) [![Build Status](https://travis-ci.org/kenshaw/secure.svg)](https://travis-ci.org/kenshaw/secure)

`secure` is an HTTP middleware for Go that facilitates some quick security wins.
It's a standard [net/http.Handler](http://golang.org/pkg/net/http/#Handler),
and can be used with many [frameworks](#integration-examples) or directly with
Go's `net/http` package.

## Installation

Install in the standard Go way:

```sh
$ go get -u github.com/kenshaw/secure
```

## Usage

```go
// examples/std/main.go
package main

import (
	"net/http"

	"github.com/kenshaw/secure"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
})

func main() {
	secureMiddleware := &secure.Middleware{
		AllowedHosts:             []string{"example.com", "ssl.example.com"},
		HostsProxyHeaders:        []string{"X-Forwarded-Host"},
		SSLRedirect:              true,
		SSLHost:                  "ssl.example.com",
		SSLForwardedProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:               315360000,
		STSIncludeSubdomains:     true,
		STSPreload:               true,
		FrameDeny:                true,
		ContentTypeNosniff:       true,
		BrowserXSSFilter:         true,
		ContentSecurityPolicy:    "default-src 'self'",
	}

	app := secureMiddleware.Handler(myHandler)
	http.ListenAndServe("127.0.0.1:3000", app)
}
```

Be sure to include the Secure middleware as close to the top (beginning) as
possible (but after logging and recovery). It's best to do the allowed hosts
and SSL check before anything else.

The above example will only allow requests with a host name of 'example.com',
or 'ssl.example.com'. Also if the request is not HTTPS, it will be redirected
to HTTPS with the host name of 'ssl.example.com'.

Once those requirements are satisfied, it will add the following headers:

```http
Strict-Transport-Security: 315360000; includeSubdomains; preload
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

### Set the `DevEnvironment` option to `true` when developing!

When `DevEnvironment` is true, the AllowedHosts, SSLRedirect, STS header, and
HPKP header will not be in effect. This allows you to work in development/test
mode and not have any annoying redirects to HTTPS (ie. development can happen
over HTTP), or block `localhost` has a bad host.

### Configuration

Secure comes with a variety of configuration options that can be set either
directly on the `secure.Middleware` type, or by using the functional option
pattern via a call to `secure.New`. Please see the [GoDoc](https://godoc.org/github.com/kenshaw/secure)
listing for more information.

### Default options

These are the preset options for Secure:

### Redirecting HTTP to HTTPS

If you want to redirect all HTTP requests to HTTPS, you can use the following
example.

```go
// examples/redirect/main.go
package main

import (
	"log"
	"net/http"

	"github.com/kenshaw/secure"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
})

func main() {
	secureMiddleware := &secure.Middleware{
		SSLRedirect: true,

		// This is optional in production. The default behavior is to just
		// redirect the request to the HTTPS protocol. Example:
		// http://github.com/some_page would be redirected to
		// https://github.com/some_page.
		SSLHost: "localhost:8443",
	}

	app := secureMiddleware.Handler(myHandler)

	// HTTP
	go func() {
		log.Fatal(http.ListenAndServe(":8080", app))
	}()

	// HTTPS
	// To generate a development cert and key, run the following from your *nix terminal:
	// go run $GOROOT/src/crypto/tls/generate_cert.go --host="localhost"
	log.Fatal(http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", app))
}
```

### Strict Transport Security

The STS header will only be sent on verified HTTPS connections (and when
`DevEnvironment` is false). Be sure to set the `SSLProxyHeaders` option if your
application is behind a proxy to ensure the proper behavior. If you need the
STS header for all HTTP and HTTPS requests (which you [shouldn't](http://tools.ietf.org/html/rfc6797#section-7.2)),
you can use the `ForceSTSHeader` option. Note that if `DevEnvironment` is true,
it will still disable this header even when `ForceSTSHeader` is set to true.

* The `preload` flag is required for domain inclusion in Chrome's [preload](https://hstspreload.appspot.com/) list.

### Content Security Policy

If you need dynamic support for CSP while using Websockets, check out this
other middleware [awakenetworks/csp](https://github.com/awakenetworks/csp).

## Integration examples

### [chi](https://github.com/pressly/chi)
```go
// examples/chi/main.go
package main

import (
	"net/http"

	"github.com/kenshaw/secure"
	"github.com/pressly/chi"
)

func main() {
	secureMiddleware := &secure.Middleware{
		FrameDeny: true,
	}

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("X-Frame-Options header is now `DENY`."))
	})
	r.Use(secureMiddleware.Handler)

	http.ListenAndServe("127.0.0.1:3000", r)
}
```

### [Echo](https://github.com/labstack/echo)
```go
// examples/echo/main.go
package main

import (
	"net/http"

	"github.com/kenshaw/secure"
	"github.com/labstack/echo"
)

func main() {
	secureMiddleware := &secure.Middleware{
		FrameDeny: true,
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "X-Frame-Options header is now `DENY`.")
	})

	e.Use(echo.WrapMiddleware(secureMiddleware.Handler))
	e.Logger.Fatal(e.Start("127.0.0.1:3000"))
}
```

### [Gin](https://github.com/gin-gonic/gin)
```go
// examples/gin/main.go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kenshaw/secure"
)

func main() {
	secureMiddleware := &secure.Middleware{
		FrameDeny: true,
	}
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := secureMiddleware.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				c.Abort()
				return
			}

			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				c.Abort()
			}
		}
	}()

	router := gin.Default()
	router.Use(secureFunc)

	router.GET("/", func(c *gin.Context) {
		c.String(200, "X-Frame-Options header is now `DENY`.")
	})

	router.Run("127.0.0.1:3000")
}
```

### [Goji](https://github.com/zenazn/goji)
```go
// examples/goji/main.go
package main

import (
	"net/http"

	"github.com/kenshaw/secure"
	"goji.io"
	"goji.io/pat"
)

func main() {
	mux := goji.NewMux()
	mux.Use(secure.New(
		secure.FrameDeny(true),
	).Handler)

	mux.HandleFunc(pat.Get("/"), func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("X-Frame-Options header is now `DENY`."))
	})

	http.ListenAndServe(":8080", mux)
}
```

### [Iris](https://github.com/kataras/iris)
```go
// examples/iris/main.go
package main

import (
	"github.com/kataras/iris"
	"github.com/kenshaw/secure"
)

func main() {
	secureMiddleware := &secure.Middleware{
		FrameDeny: true,
	}

	iris.UseFunc(func(c *iris.Context) {
		err := secureMiddleware.Process(c.ResponseWriter, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	})

	iris.Get("/home", func(c *iris.Context) {
		c.SendStatus(200, "X-Frame-Options header is now `DENY`.")
	})

	iris.Listen(":8080")
}
```

### [Negroni](https://github.com/codegangsta/negroni)

Note that the `secure.Middleware` type has a special helper function
`HandlerFuncWithNext` for use with Negroni.

```go
// examples/negroni/main.go
package main

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/kenshaw/secure"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("X-Frame-Options header is now `DENY`."))
	})

	secureMiddleware := &secure.Middleware{
		FrameDeny: true,
	}

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	n.UseHandler(mux)

	n.Run("127.0.0.1:3000")
}
```

## nginx

If you would like to add the above security rules directly to your
[nginx](http://wiki.nginx.org/Main) configuration, everything is below:

```nginx
# Allowed Hosts:
if ($host !`* ^(example.com|ssl.example.com)$ ) {
    return 500;
}

# SSL Redirect:
server {
    listen      80;
    server_name example.com ssl.example.com;
    return 301 https://ssl.example.com$request_uri;
}

# Headers to be added:
add_header Strict-Transport-Security "max-age=315360000";
add_header X-Frame-Options "DENY";
add_header X-Content-Type-Options "nosniff";
add_header X-XSS-Protection "1; mode=block";
add_header Content-Security-Policy "default-src 'self'";
```
