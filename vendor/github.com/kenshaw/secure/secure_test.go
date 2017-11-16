package secure

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bar"))
})

func TestNoConfig(t *testing.T) {
	s := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")
}

func TestEmptyAllowedHosts(t *testing.T) {
	s := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestOneAllowedHost(t *testing.T) {
	s := New(AllowedHosts("www.example.com"))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadAllowedHost(t *testing.T) {
	s := New(AllowedHosts("sub.example.com"))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestGoodSingleAllowHostsProxyHeaders(t *testing.T) {
	s := New(
		AllowedHosts("www.example.com"),
		HostsProxyHeaders("X-Proxy-Host"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.Header.Set("X-Proxy-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadSingleAllowHostsProxyHeaders(t *testing.T) {
	s := New(
		AllowedHosts("sub.example.com"),
		HostsProxyHeaders("X-Proxy-Host"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.Header.Set("X-Proxy-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestGoodMultipleAllowHosts(t *testing.T) {
	s := New(
		AllowedHosts("www.example.com", "sub.example.com"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "sub.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)
}

func TestBadMultipleAllowHosts(t *testing.T) {
	s := New(
		AllowedHosts("www.example.com", "sub.example.com"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}

func TestAllowHostsInDevMode(t *testing.T) {
	s := New(
		AllowedHosts("www.example.com", "sub.example.com"),
		DevEnvironment(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestBadHostHandler(t *testing.T) {
	s := New(
		AllowedHosts("www.example.com", "sub.example.com"),
		BadHostHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "BadHost", http.StatusInternalServerError)
		}),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	// http.Error outputs a new line character with the response.
	expect(t, res.Body.String(), "BadHost\n")
}

func TestSSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestSSLInDevMode(t *testing.T) {
	s := New(
		SSLRedirect(true),
		DevEnvironment(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestBasicSSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestBasicSSLWithHost(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLHost("secure.example.com"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestBadProxySSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestCustomProxySSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "https"}),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomProxySSLInDevMode(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "https"}),
		DevEnvironment(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "http")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomProxyAndHostProxyHeadersWithRedirect(t *testing.T) {
	s := New(
		HostsProxyHeaders("X-Forwarded-Host"),
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "http"}),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "example-internal"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")
	req.Header.Add("X-Forwarded-Host", "www.example.com")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://www.example.com/foo")
}

func TestCustomProxyAndHostSSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "https"}),
		SSLHost("secure.example.com"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
}

func TestCustomBadProxyAndHostSSL(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "superman"}),
		SSLHost("secure.example.com"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusMovedPermanently)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestCustomBadProxyAndHostSSLWithTempRedirect(t *testing.T) {
	s := New(
		SSLRedirect(true),
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "superman"}),
		SSLHost("secure.example.com"),
		SSLTemporaryRedirect(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www.example.com"
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusTemporaryRedirect)
	expect(t, res.Header().Get("Location"), "https://secure.example.com/foo")
}

func TestSTSHeaderWithNoSSL(t *testing.T) {
	s := New(
		STSSeconds(315360000),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestSTSHeaderWithNoSSLButWithForce(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		ForceSTSHeader(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000")
}

func TestSTSHeaderWithNoSSLButWithForceAndDev(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		ForceSTSHeader(true),
		DevEnvironment(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestSTSHeaderWithSSL(t *testing.T) {
	s := New(
		SSLForwardedProxyHeaders(map[string]string{"X-Forwarded-Proto": "https"}),
		STSSeconds(315360000),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "http"
	req.Header.Add("X-Forwarded-Proto", "https")

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000")
}

func TestSTSHeaderInDevMode(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		DevEnvironment(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "")
}

func TestSTSHeaderWithSubdomains(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		STSIncludeSubdomains(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; includeSubdomains")
}

func TestSTSHeaderWithPreload(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		STSPreload(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; preload")
}

func TestSTSHeaderWithSubdomainsWithPreload(t *testing.T) {
	s := New(
		STSSeconds(315360000),
		STSIncludeSubdomains(true),
		STSPreload(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.URL.Scheme = "https"

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Strict-Transport-Security"), "max-age=315360000; includeSubdomains; preload")
}

func TestFrameDeny(t *testing.T) {
	s := New(
		FrameDeny(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "DENY")
}

func TestCustomFrameValue(t *testing.T) {
	s := New(
		CustomFrameOptionsValue("SAMEORIGIN"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "SAMEORIGIN")
}

func TestCustomFrameValueWithDeny(t *testing.T) {
	s := New(
		FrameDeny(true),
		CustomFrameOptionsValue("SAMEORIGIN"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "SAMEORIGIN")
}

func TestContentNosniff(t *testing.T) {
	s := New(
		ContentTypeNosniff(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Content-Type-Options"), "nosniff")
}

func TestXSSProtection(t *testing.T) {
	s := New(
		BrowserXSSFilter(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-XSS-Protection"), "1; mode=block")
}

func TestCustomXSSProtection(t *testing.T) {
	xssVal := "1; report=https://example.com"
	s := New(
		CustomBrowserXSSValue(xssVal),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-XSS-Protection"), xssVal)
}

func TestBothXSSProtection(t *testing.T) {
	xssVal := "0"
	s := New(
		BrowserXSSFilter(true),
		CustomBrowserXSSValue(xssVal),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-XSS-Protection"), xssVal)
}

func TestContentSecurityPolicy(t *testing.T) {
	s := New(
		ContentSecurityPolicy("default-src 'self'"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestInlineSecure(t *testing.T) {
	s := New(
		FrameDeny(true),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.HandlerFuncWithNext(w, r, nil)
		w.Write([]byte("bar"))
	})

	handler.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("X-Frame-Options"), "DENY")
}

func TestReferrer(t *testing.T) {
	s := New(
		ReferrerPolicy("same-origin"),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get("Referrer-Policy"), "same-origin")
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
