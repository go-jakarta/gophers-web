// gophers-web is the web server for gophers.id.
package main

//go:generate assetgen

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/brankas/envcfg"
	"github.com/brankas/stringid"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/csrf"
	"github.com/kenshaw/logrusmw"
	"github.com/kenshaw/secure"
	"github.com/leonelquinteros/gotext"
	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/shurcooL/httpgzip"
	"github.com/sirupsen/logrus"
	"github.com/tylerb/graceful"
	"goji.io"
	"goji.io/pat"
	"goji.io/pattern"

	// assets
	"gophers.id/gophers-web/assets"
	"gophers.id/gophers-web/assets/geoip"
	"gophers.id/gophers-web/assets/locales"
	"gophers.id/gophers-web/assets/templates"
)

// general config and flags
var (
	// values that can be baked in via build
	env      = "development" // environment
	isDevEnv = true

	// globals
	logger *logrus.Logger // master log

	// config
	config *envcfg.Envcfg

	// geoip
	geoipdb *maxminddb.Reader

	// translations
	translations map[string]*gotext.Po

	// indexPage
	indexPage templates.Page
)

func init() {
	var err error

	// setup logrus
	logger = logrus.New()

	// force all writes to regular log to logger
	log.SetOutput(logger.Writer())
	log.SetFlags(0)

	// load config variables from $ENV{APP_CONFIG}
	config, err = envcfg.New()
	if err != nil {
		logger.Fatal(err)
	}

	// setup environment
	env = config.GetKey("runtime.environment")
	isDevEnv = env == "development"

	// change environment for development
	if isDevEnv {
		tf := new(logrus.TextFormatter)
		//tf.ForceColors = logrus.IsTerminal()
		tf.FullTimestamp = true

		logger.Formatter = tf
	}

	// init geoip data
	geoipdb, err = initGeoip()
	if err != nil {
		logger.Fatal(err)
	}

	// update every hour the event information
	mapsToken := config.GetKey("google.mapstoken")
	indexPage = &templates.IndexPage{GoogleMapsToken: mapsToken}
	go func() {
		time.Sleep(1 * time.Hour)
	}()
}

func main() {
	var err error

	err = setupAssets()
	if err != nil {
		logger.Fatal(err)
	}

	server := setupServer()
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}

}

func initGeoip() (*maxminddb.Reader, error) {
	var err error

	// load geoip data
	gz, err := loadasset("GeoLite2-Country.mmdb.gz", geoip.Geoip)
	if err != nil {
		return nil, err
	}

	// decompress geoip data
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return maxminddb.FromBytes(buf)
}

// setup static assets
func setupAssets() error {
	// create locales storage
	translations = map[string]*gotext.Po{
		"en": {},
		"id": {},
	}

	// load translations
	for l, po := range translations {
		buf, err := loadasset(l+".po", locales.Locales)
		if err != nil {
			return err
		}
		po.Parse(buf)
	}

	logger.Printf("processed %d translations", len(translations))

	// set template stuff
	templates.Asset = func(fn string) string {
		return "/_/" + assets.ManifestPath("/"+strings.TrimPrefix(fn, "/"))
	}
	templates.CsrfToken = func(req *http.Request) string {
		return string(csrf.TemplateField(req))
	}

	return nil
}

type keyType int

const (
	loggerKey keyType = iota
	countryKey
)

// setup server
func setupServer() *graceful.Server {
	logger.Infof("initializing middleware (environment: %s)", env)

	mux := goji.NewMux()

	// add logger
	mux.Use(stringid.Middleware())
	mux.Use(func(next http.Handler) http.Handler {
		mw := func(res http.ResponseWriter, req *http.Request) {
			if !isDevEnv {
				// set remote addr to x-forwarded-for
				if addr := req.Header.Get("X-Forwarded-For"); addr != "" {
					req.RemoteAddr = addr
				}
			}

			ctxt := req.Context()

			// extract locale
			lang, country := extractLocale(res, req)

			// create logger fields
			fields := logrus.Fields{
				"req_id": stringid.FromContext(ctxt),
				"l":      lang,
				"loc":    country,
			}
			if !isDevEnv {
				fields["req"] = req
			}

			// add values to context
			ctxt = context.WithValue(ctxt, loggerKey, logger.WithFields(fields))
			ctxt = context.WithValue(ctxt, countryKey, country)
			ctxt = context.WithValue(ctxt, templates.LanguageKey, lang)

			next.ServeHTTP(res, req.WithContext(ctxt))
		}

		return http.HandlerFunc(mw)
	})
	mux.Use(logrusmw.NewWithID(logger, stringid.FromContext))

	// add basic security options
	mux.Use(secure.New(
		secure.AllowedHosts(
			"gophers.id", "www.gophers.id",
		),
		secure.SSLRedirect(true),
		secure.SSLHost("gophers.id"),
		secure.SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "https"}),
		secure.STSSeconds(315360000),
		secure.STSIncludeSubdomains(true),
		secure.STSPreload(true),
		secure.FrameDeny(true),
		secure.ContentTypeNosniff(true),
		secure.BrowserXSSFilter(true),
		//secure.ContentSecurityPolicy: "default-src 'self'", // TODO: fix this
		secure.DevEnvironment(isDevEnv), // toggle development depending on environment
		secure.BadHostHandler(func(res http.ResponseWriter, req *http.Request) {
			http.Redirect(res, req, "https://gophers.id/", http.StatusMovedPermanently)
		}),
	).Handler)

	// add gorilla/csrf middleware
	mux.Use(csrf.Protect(
		[]byte(config.GetKey("server.csrftoken")),
		csrf.RequestHeader("X-CSRF"),
		csrf.CookieName("__csrf"),
		csrf.FieldName("__csrf"),
		csrf.Secure(!isDevEnv),
		csrf.ErrorHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			http.Error(res, "invalid csrf token", http.StatusForbidden)
		})),
	))

	// add gzip compression
	mux.Use(gziphandler.GzipHandler)

	// add static assets
	mux.Handle(pat.New("/_/*"), assets.StaticHandler(pattern.Path))

	// add template pages
	idTrans := translator("id")
	transMap := map[string]templates.TransFunc{
		"id": idTrans,
		"en": translator("en"),
		"es": translator("es"),
	}

	// robots.txt
	mux.HandleFunc(pat.Get("/robots.txt"), func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(res, robotsTxt)
	})

	mux.HandleFunc(pat.Get("/*"), func(res http.ResponseWriter, req *http.Request) {
		// handle the go-get stuff
		if req.URL.Query().Get("go-get") == "1" {
			repo := strings.TrimPrefix(req.URL.Path, "/")
			if i := strings.Index(repo, "/"); i != -1 {
				repo = repo[:i]
			}

			// determine the repository
			repo = strings.TrimSuffix(repo, ".git")
			url := "https://github.com/go-jakarta/" + repo + ".git"
			if strings.ToLower(repo) == "gojakarta" {
				url = "https://github.com/go-jakarta/slides.git"
			}

			fmt.Fprintf(res, gitHTML, `gophers.id/`+repo+` git `+url, url, url)
			return
		}

		// do a redirect ...
		reqPath := strings.ToLower(strings.TrimPrefix(req.URL.Path, "/"))
		switch reqPath {
		case "gojakarta":
			http.Redirect(res, req, "https://meetup.com/GoJakarta", http.StatusMovedPermanently)

		case "slides":
			http.Redirect(res, req, "https://github.com/go-jakarta/slides", http.StatusMovedPermanently)

		case "index", "index.html", "":
			templates.Do(res, req, indexPage, idTrans, transMap)

		default:
			http.NotFound(res, req)
		}
	})

	// setup graceful
	return &graceful.Server{
		Server: &http.Server{
			Addr:    ":" + config.GetKey("server.port"),
			Handler: mux,
		},
		Timeout: 10 * time.Second,
		BeforeShutdown: func() bool {
			logger.Infoln("received signal")
			return true
		},
		ShutdownInitiated: func() {
			logger.Infoln("done.")
		},
	}
}

// lookupCountry returns the geoip country.
func lookupCountry(remoteAddr string) string {
	host, _, _ := net.SplitHostPort(remoteAddr)
	a := net.ParseIP(host)
	if a == nil {
		return "id"
	}

	var rec struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	err := geoipdb.Lookup(a, &rec)
	if err != nil || rec.Country.ISOCode == "" {
		return "id"
	}

	return strings.ToLower(rec.Country.ISOCode)
}

var langRE = regexp.MustCompile("^(en|id|es)$")

const (
	langPrefCookie = "lang-pref"
)

// extractLocale extracts the language and country from the request.
func extractLocale(res http.ResponseWriter, req *http.Request) (string, string) {
	country := lookupCountry(req.RemoteAddr)

	// ?l=lang was passed, set cookie, and return
	if l := req.URL.Query().Get("l"); l != "" && langRE.MatchString(l) {
		http.SetCookie(res, &http.Cookie{
			Name:    langPrefCookie,
			Value:   l,
			Expires: time.Now().Add(7 * 24 * time.Hour),
		})

		return l, country
	}

	// read lang-pref cookie
	if c, err := req.Cookie(langPrefCookie); err == nil && langRE.MatchString(c.Value) {
		return c.Value, country
	}

	// read Accept-Language header
	if accept := header.ParseAccept(req.Header, "Accept-Language"); len(accept) > 0 {
		for _, a := range accept {
			l := strings.ToLower(a.Value)
			if i := strings.Index(l, "-"); i != -1 {
				l = l[:i]
			}

			if _, ok := translations[l]; ok && (country != "id" || l != "en") {
				return l, country
			}
		}
	}

	// if nothing set, and remote address is not in ID
	if country != "id" {
		return "en", country
	}

	return "id", country
}

func translator(lang string) templates.TransFunc {
	get := translations[lang].Get

	if lang == "en" {
		return fmt.Sprintf
	}

	return func(s string, v ...interface{}) string {
		if x := get(s, v...); x != "" {
			return x
		}

		return fmt.Sprintf(s, v...)
	}
}

const (
	gitHTML = `<!DOCTYPE html>
<html>
<head>
  <meta name="go-import" content="%s">
</head>
<body>
  <a href="%s">%s</a>
</body>
</html>
`

	robotsTxt = `User-agent: *
Allow: *
`
)

func loadasset(fn string, fs http.FileSystem) ([]byte, error) {
	fn = "/" + strings.TrimPrefix(fn, "/")

	f, err := fs.Open(fn)
	if err != nil {
		return nil, err
	}

	switch x := f.(type) {
	case httpgzip.GzipByter:
		gz, err := gzip.NewReader(bytes.NewReader(x.GzipBytes()))
		if err != nil {
			return nil, fmt.Errorf("unable to decode asset %s", fn)
		}
		return ioutil.ReadAll(gz)

	case httpgzip.NotWorthGzipCompressing:
		return ioutil.ReadAll(f)
	}
	return nil, fmt.Errorf("unknown type for asset %s", fn)
}
