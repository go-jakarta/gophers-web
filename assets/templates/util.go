package templates

import (
	"net/http"
	"time"

	qtpl "github.com/valyala/quicktemplate"
)

type AssetFunc func(path string) string
type TransFunc func(s string, v ...interface{}) string
type CsrfTokenFunc func(req *http.Request) string

var (
	Asset     AssetFunc
	CsrfToken CsrfTokenFunc

	year string
)

type keyType int

const (
	LanguageKey keyType = iota
)

func init() {
	year = time.Now().Format("2006")

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			year = time.Now().Format("2006")
		}
	}()
}

// NewHandler creates a template handler for the page.
func NewHandler(p Page, def TransFunc, translations map[string]TransFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctxt := req.Context()
		w := qtpl.AcquireWriter(res)
		defer qtpl.ReleaseWriter(w)

		// grab translator from context
		lang := ctxt.Value(LanguageKey).(string)
		T, ok := translations[lang]
		if !ok {
			T = def
		}

		StreamLayoutPage(w, p, req, lang, T)
	}
}

var translations = []struct {
	lang, name string
}{
	0: {"id", "Bahasa Indonesia"},
	1: {"en", "English"},
}
