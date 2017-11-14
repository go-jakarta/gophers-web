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

	jakarta *time.Location
)

type keyType int

const (
	LanguageKey keyType = iota
)

func init() {
	time.Local = time.UTC

	var err error
	jakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}

	year = time.Now().In(jakarta).Format("2006")

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			year = time.Now().In(jakarta).Format("2006")
		}
	}()
}

// Do displays the page using the supplied translation func and translations.
func Do(res http.ResponseWriter, req *http.Request, p Page, def TransFunc, translations map[string]TransFunc) {
	w := qtpl.AcquireWriter(res)
	defer qtpl.ReleaseWriter(w)

	// grab translator from context
	lang := req.Context().Value(LanguageKey).(string)
	T, ok := translations[lang]
	if !ok {
		T = def
	}

	StreamLayoutPage(w, p, req, lang, T)
}

var translations = []struct {
	lang, name string
}{
	0: {"id", "Bahasa Indonesia"},
	1: {"en", "English"},
}
