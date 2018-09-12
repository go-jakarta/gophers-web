package wutil

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"goji.io"
	"goji.io/pat"
	"goji.io/pattern"
)

const (
	// DefaultCsrfVariableName is the default csrf variable name used in asset
	// sets.
	DefaultCsrfVariableName = "__csrf_token__"

	// DefaultAssetPath is the default asset path prefix for static assets.
	DefaultAssetPath = "/_/"

	// DefaultManifestPath is the path to the manifest in the asset set to use
	// to determine what assets to serve.
	DefaultManifestPath = "manifest.json"

	// DefaultFaviconPath is the path to the favicon.
	DefaultFaviconPath = "favicon.ico"

	// DefaultTemplatesSuffix is the default suffix to use for templates.
	DefaultTemplatesSuffix = ".html"
)

// lgr is the common interface used for logging.
type lgr func(string, ...interface{})

type assetNameFn func() []string
type assetFn func(string) ([]byte, error)
type assetInfoFn func(string) (os.FileInfo, error)

type asset struct {
	info        *os.FileInfo
	data        *[]byte
	sha1        string
	contentType string
}

// AssetSet is a collection of static assets.
type AssetSet struct {
	// retrieval functions
	assetNameFn assetNameFn
	assetFn     assetFn
	assetInfoFn assetInfoFn

	// csrfVariableName is the name to add to the template context for the
	// generate csrf token.
	csrfVariableName string

	// csrf is a func that returns the csrf token for the request.
	csrf func(*http.Request) string

	// assetPath is the location of the bin'd assets.
	assetPath string

	// manifestPath is the path of the manifest file in the asset set.
	manifestPath string

	// faviconPath is the path to the favicon in the asset set.
	faviconPath string

	// manifest is the map of the original asset name -> hashed.
	manifest map[string]string

	// processed asset data
	assets map[string]*asset

	// logger
	logger lgr

	// ignore
	ignore []*regexp.Regexp
}

// AssetSetOption represents options when creating a new AssetSet.
type AssetSetOption func(*AssetSet)

// Path sets the root lookup path for the AssetSet.
func Path(path string) AssetSetOption {
	return func(as *AssetSet) {
		as.assetPath = path
	}
}

// ManifestPath changes the default manifest path in the AssetSet.
func ManifestPath(path string) AssetSetOption {
	return func(as *AssetSet) {
		as.manifestPath = path
	}
}

// FaviconPath changes the default favicon path in the AssetSet.
func FaviconPath(path string) AssetSetOption {
	return func(as *AssetSet) {
		as.faviconPath = path
	}
}

// Ignore prevents files matching the supplied regexps to be excluded from
// being served from the AssetSet.
func Ignore(regexps ...*regexp.Regexp) AssetSetOption {
	return func(as *AssetSet) {
		as.ignore = append(as.ignore, regexps...)
	}
}

// Logger sets the logger for an AssetSet.
func Logger(l lgr) AssetSetOption {
	return func(as *AssetSet) {
		as.logger = l
	}
}

// CsrfVariableName sets the csrf variable name used in the AssetSet.
func CsrfVariableName(name string) AssetSetOption {
	return func(as *AssetSet) {
		as.csrfVariableName = name
	}
}

// Csrf sets the func that generates a csrf token from the context for an
// AssetSet.
//
// Templates will have this value available to them via as {{ csrf }}.
func Csrf(f func(*http.Request) string) AssetSetOption {
	return func(as *AssetSet) {
		as.csrf = f
	}
}

// NewAssetSet creates an asset set with the passed parameters.
func NewAssetSet(anFn assetNameFn, aFn assetFn, aiFn assetInfoFn, opts ...AssetSetOption) (*AssetSet, error) {
	as := &AssetSet{
		assetNameFn: anFn,
		assetFn:     aFn,
		assetInfoFn: aiFn,

		csrfVariableName: DefaultCsrfVariableName,

		csrf: nil,

		assetPath:    DefaultAssetPath,
		manifestPath: DefaultManifestPath,
		faviconPath:  DefaultFaviconPath,

		logger: log.Printf,
		ignore: []*regexp.Regexp{},

		manifest: make(map[string]string),
		assets:   make(map[string]*asset),
	}

	// apply options
	for _, o := range opts {
		o(as)
	}

	if !strings.HasSuffix(as.assetPath, "/") {
		return nil, errors.New("asset path must end with /")
	}

	// grab manifest bytes
	mfd, err := aFn(as.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("could not read data from manifest '%s'", as.manifestPath)
	}

	// load manifest data
	var mf interface{}
	err = json.Unmarshal(mfd, &mf)
	if err != nil {
		return nil, fmt.Errorf("could not json.Unmarshal manifest '%s': %s", as.manifestPath, err)
	}

	// convert mf to actual map
	manifestMap, ok := mf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'%s' is not a valid manifest", as.manifestPath)
	}

	// process static assets
	ignoredCount := 0
	for name, v := range manifestMap {
		ignored := false

		// determine if the name is in the ignore list
		for _, re := range as.ignore {
			if re.MatchString(name) {
				ignored = true
				ignoredCount++
				break
			}
		}

		// only process if the asset is not on ignore list
		if !ignored {
			hash, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value for key '%s' in manifest '%s'", name, as.manifestPath)
			}

			data, err := aFn(name)
			if err != nil {
				return nil, fmt.Errorf("asset %s (%s) not available", name, hash)
			}

			info, err := aiFn(name)
			if err != nil {
				return nil, fmt.Errorf("asset info %s (%s) not available", name, hash)
			}

			// store data
			as.manifest[name] = hash
			as.assets[hash] = &asset{
				info:        &info,
				data:        &data,
				sha1:        fmt.Sprintf("%x", sha1.Sum(data)),
				contentType: as.contentType(name),
			}
		}
	}

	// format ignored
	ignoredStr := ""
	if ignoredCount > 0 {
		ignoredStr = fmt.Sprintf(", ignored: %d", ignoredCount)
	}

	as.logger("processed static assets (%d%s)", len(as.manifest), ignoredStr)

	return as, nil
}

// staticHandler retrieves the static asset from the AssetSet, sending it to
// the http endpoint.
func (as *AssetSet) staticHandler(name string, res http.ResponseWriter, req *http.Request) {
	// grab info
	assetItem, ok := as.assets[name]
	if !ok {
		http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// grab modtime
	modtime := (*assetItem.info).ModTime()

	// check if-modified-since header, bail if present
	if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && modtime.Unix() <= t.Unix() {
		res.WriteHeader(http.StatusNotModified) // 304
		return
	}

	// check If-None-Match header, bail if present and match sha1
	if req.Header.Get("If-None-Match") == assetItem.sha1 {
		res.WriteHeader(http.StatusNotModified) // 304
		return
	}

	// set headers
	res.Header().Set("Content-Type", assetItem.contentType)
	res.Header().Set("Date", time.Now().Format(http.TimeFormat))

	// cache headers
	res.Header().Set("Cache-Control", "public, no-transform, max-age=31536000")
	res.Header().Set("Expires", time.Now().AddDate(1, 0, 0).Format(http.TimeFormat))
	res.Header().Set("Last-Modified", modtime.Format(http.TimeFormat))
	res.Header().Set("ETag", assetItem.sha1)

	// write data to response
	res.Write(*(assetItem.data))
}

// contentType returns the content type based on a file's name.
func (as *AssetSet) contentType(name string) string {
	// determine content type
	typ := "application/octet-stream"
	pos := strings.LastIndex(name, ".")
	if pos >= 0 {
		typ = mime.TypeByExtension(name[pos:])
	}

	return typ
}

// StaticHandler serves static assets from the AssetSet.
func (as *AssetSet) StaticHandler(res http.ResponseWriter, req *http.Request) {
	as.staticHandler(pattern.Path(req.Context())[1:], res, req)
}

// FaviconHandler is a helper that serves the static "favicon.ico" asset from
// the AssetSet.
func (as *AssetSet) FaviconHandler(res http.ResponseWriter, req *http.Request) {
	as.staticHandler(as.faviconPath, res, req)
}

// AssetPath retrieves the asset's manifest path and returns the path prefix
// along with the path hash.
func (as *AssetSet) AssetPath(path string) string {
	// load the path from the manifest and return if valid
	p, ok := as.manifest[path]
	if !ok {
		// asset not in manifest
		as.logger("asset %s not found in manifest", path)
		return "NA"
	}

	return as.assetPath + p
}

// Register registers the AssetSet to the provided mux.
func (as *AssetSet) Register(mux *goji.Mux) {
	// add favicon handler only if the favicon.ico is present in the path.
	if false {
		mux.HandleFunc(pat.Get("/favicon.ico"), as.FaviconHandler)
	}

	mux.HandleFunc(pat.Get(as.assetPath+"*"), as.StaticHandler)
}

func init() {
	mime.AddExtensionType("ico", "image/x-icon")
	mime.AddExtensionType("woff2", "application/font-woff2")
}
