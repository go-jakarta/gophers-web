// gophers-web is the web server for gophers.id.
package main

//go:generate go run gen.go

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

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/csrf"
	"github.com/kenshaw/glogrus2"
	"github.com/kenshaw/gojiid"
	"github.com/knq/envcfg"
	"github.com/knq/wutil"
	"github.com/leonelquinteros/gotext"
	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/sirupsen/logrus"
	"github.com/tylerb/graceful"
	"github.com/unrolled/secure"

	// goji
	"goji.io"
	"goji.io/pat"

	// assets
	wwwdata "gophers.id/gophers-web/assets"
	wwwgeo "gophers.id/gophers-web/assets/geoip"
	wwwloc "gophers.id/gophers-web/assets/locales"
	wwwtpl "gophers.id/gophers-web/assets/templates"
)

// general config and flags
var (
	// values that can be baked in via build
	env      = "development" // environment
	isDevEnv = true

	// globals
	logger *logrus.Logger // master log

	// assets
	assetSet *wutil.AssetSet

	// config
	config *envcfg.Envcfg

	// google service account credentials
	googleCreds []byte

	// geoip
	geoip *maxminddb.Reader

	// translations
	translations map[string]*gotext.Po
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

	// read google creds
	googleCreds = []byte(config.GetKey("google.creds"))

	// change environment for development
	if isDevEnv {
		tf := new(logrus.TextFormatter)
		//tf.ForceColors = logrus.IsTerminal()
		tf.FullTimestamp = true

		logger.Formatter = tf
	} else {
		// add stackdriver hook for logrus
		/*
			hook, err := sdhook.New(
				sdhook.GoogleServiceAccountCredentialsJSON(googleCreds),
				sdhook.LogName("gophers-web"),
			)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Hooks.Add(hook)
		*/
	}

	// init geoip data
	geoip, err = initGeoIP()
	if err != nil {
		logger.Fatal(err)
	}
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

func initGeoIP() (*maxminddb.Reader, error) {
	var err error

	// load geoip data
	gz, err := wwwgeo.Asset("GeoLite2-Country.mmdb.gz")
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
	var err error

	opts := []wutil.AssetSetOption{
		wutil.Logger(logger.Printf),
		wutil.Csrf(func(req *http.Request) string {
			return string(csrf.TemplateField(req))
		}),
	}
	if !isDevEnv {
		opts = append(opts, wutil.Ignore(regexp.MustCompile(`\.map$`)))
	}

	// create asset set
	assetSet, err = wutil.NewAssetSet(
		wwwdata.AssetNames,
		wwwdata.Asset,
		wwwdata.AssetInfo,
		opts...,
	)
	if err != nil {
		return err
	}

	// create locales storage
	translations = map[string]*gotext.Po{
		"en": &gotext.Po{},
		"id": &gotext.Po{},
		"es": &gotext.Po{},
	}

	// load translations
	for l, po := range translations {
		buf, err := wwwloc.Asset(l + ".po")
		if err != nil {
			return err
		}

		po.Parse(string(buf))
	}

	logger.Printf("processed %d translations", len(translations))

	// set template stuff
	wwwtpl.Asset = assetSet.AssetPath
	wwwtpl.CsrfToken = func(req *http.Request) string {
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
	mux.Use(gojiid.NewRequestId())
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
				"req_id": gojiid.FromContext(ctxt),
				"l":      lang,
				"loc":    country,
			}
			if !isDevEnv {
				fields["req"] = req
			}

			// add values to context
			ctxt = context.WithValue(ctxt, loggerKey, logger.WithFields(fields))
			ctxt = context.WithValue(ctxt, countryKey, country)
			ctxt = context.WithValue(ctxt, wwwtpl.LanguageKey, lang)

			next.ServeHTTP(res, req.WithContext(ctxt))
		}

		return http.HandlerFunc(mw)
	})
	mux.Use(glogrus2.NewWithReqId(logger, "gophers-web", gojiid.FromContext))

	// add basic security options
	mux.Use(secure.New(secure.Options{
		AllowedHosts: []string{
			"gophers.id", "www.gophers.id",
		},
		SSLRedirect:          true,
		SSLHost:              "gophers.id",
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		//ContentSecurityPolicy: "default-src 'self'", // TODO: fix this
		//PublicKey: fmt.Sprintf(""), // TODO: add this (public-key-pinning)

		// toggle development depending on environment
		IsDevelopment: isDevEnv,
	}).Handler)

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

	// add general utility handlers
	wutil.RegisterUtils(mux)

	// add static assets
	assetSet.Register(mux)

	// add template pages
	idTrans := translator("id")
	transMap := map[string]wwwtpl.TransFunc{
		"id": idTrans,
		"en": translator("en"),
		"es": translator("es"),
	}

	// robots.txt
	mux.HandleFunc(pat.Get("/robots.txt"), func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(res, robotsTxt)
	})

	indexHandler := wwwtpl.NewHandler(&wwwtpl.IndexPage{}, idTrans, transMap)
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
		if strings.ToLower(strings.TrimPrefix(req.URL.Path, "/")) == "gojakarta" {
			http.Redirect(res, req, "https://meetup.com/GoJakarta", http.StatusMovedPermanently)
			return
		}

		indexHandler.ServeHTTP(res, req)
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
	a := net.ParseIP(remoteAddr)
	if a == nil {
		return "id"
	}

	var rec struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	err := geoip.Lookup(a, &rec)
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

func translator(lang string) wwwtpl.TransFunc {
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
