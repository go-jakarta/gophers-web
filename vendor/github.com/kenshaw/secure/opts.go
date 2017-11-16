package secure

import "net/http"

// Option is a secure Middleware option.
type Option func(*Middleware)

// AllowedHosts is an option to set the allowed hosts.
func AllowedHosts(allowedHosts ...string) Option {
	return func(s *Middleware) {
		s.AllowedHosts = allowedHosts
	}
}

// HostsProxyHeaders is an option to set the host proxy headers.
func HostsProxyHeaders(hostsProxyHeaders ...string) Option {
	return func(s *Middleware) {
		s.HostsProxyHeaders = hostsProxyHeaders
	}
}

// SSLRedirect is an option to toggle ssl redirect.
func SSLRedirect(sslRedirect bool) Option {
	return func(s *Middleware) {
		s.SSLRedirect = sslRedirect
	}
}

// SSLTemporaryRedirect is an option to set the SSL temporary redirect.
func SSLTemporaryRedirect(sslTemporaryRedirect bool) Option {
	return func(s *Middleware) {
		s.SSLTemporaryRedirect = sslTemporaryRedirect
	}
}

// SSLHost is an option to set the ssl host.
func SSLHost(sslHost string) Option {
	return func(s *Middleware) {
		s.SSLHost = sslHost
	}
}

// SSLForwardedProxyHeaders is an option to set the SSL forwarded proxy headers.
func SSLForwardedProxyHeaders(m map[string]string) Option {
	return func(s *Middleware) {
		s.SSLForwardedProxyHeaders = m
	}
}

// STSSeconds is an option to set the STS seconds.
func STSSeconds(stsSeconds int64) Option {
	return func(s *Middleware) {
		s.STSSeconds = stsSeconds
	}
}

// STSIncludeSubdomains is an option to set STS include subdomains.
func STSIncludeSubdomains(stsIncludeSubdomains bool) Option {
	return func(s *Middleware) {
		s.STSIncludeSubdomains = stsIncludeSubdomains
	}
}

// STSPreload is an option to set STS preload.
func STSPreload(stsPreload bool) Option {
	return func(s *Middleware) {
		s.STSPreload = stsPreload
	}
}

// ForceSTSHeader is an option to force STS header.
func ForceSTSHeader(forceSTSHeader bool) Option {
	return func(s *Middleware) {
		s.ForceSTSHeader = forceSTSHeader
	}
}

// FrameDeny is an option to set frame deny.
func FrameDeny(frameDeny bool) Option {
	return func(s *Middleware) {
		s.FrameDeny = frameDeny
	}
}

// CustomFrameOptionsValue is an option to set custom frame options value.
func CustomFrameOptionsValue(customFrameOptionsValue string) Option {
	return func(s *Middleware) {
		s.CustomFrameOptionsValue = customFrameOptionsValue
	}
}

// ContentTypeNosniff is an option to set content type to NOSNIFF.
func ContentTypeNosniff(contentTypeNosniff bool) Option {
	return func(s *Middleware) {
		s.ContentTypeNosniff = contentTypeNosniff
	}
}

// BrowserXSSFilter is an option to set browser xss filter.
func BrowserXSSFilter(browserXSSFilter bool) Option {
	return func(s *Middleware) {
		s.BrowserXSSFilter = browserXSSFilter
	}
}

// CustomBrowserXSSValue is an option to set custom browser xss value.
func CustomBrowserXSSValue(customBrowserXSSValue string) Option {
	return func(s *Middleware) {
		s.CustomBrowserXSSValue = customBrowserXSSValue
	}
}

// ContentSecurityPolicy is an option to set the content security policy.
func ContentSecurityPolicy(contentSecurityPolicy string) Option {
	return func(s *Middleware) {
		s.ContentSecurityPolicy = contentSecurityPolicy
	}
}

// ReferrerPolicy is an option to set the referrer policy.
func ReferrerPolicy(referrerPolicy string) Option {
	return func(s *Middleware) {
		s.ReferrerPolicy = referrerPolicy
	}
}

// BadHostHandler is an option to set the bad host handler.
func BadHostHandler(badHostHandler http.HandlerFunc) Option {
	return func(s *Middleware) {
		s.BadHostHandler = badHostHandler
	}
}

// DevEnvironment is an option to set toggle development environment options.
func DevEnvironment(isDevEnvironment bool) Option {
	return func(s *Middleware) {
		s.DevEnvironment = isDevEnvironment
	}
}
