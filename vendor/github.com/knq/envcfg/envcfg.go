// Package envcfg provides a common way to load configuration variables from
// the system environment or from based on initial configuration values stored
// on disk or in a base64 encoded environment variable.
package envcfg

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/brankas/autocertdns"
	"github.com/brankas/autocertdns/godop"
	"github.com/knq/ini"
	"golang.org/x/crypto/acme/autocert"
)

const (
	// DefaultVarName is the default environment variable name to load the
	// initial configuration data from.
	DefaultVarName = "APP_CONFIG"

	// DefaultConfigFile is the default file path to load the initial
	// configuration data from.
	DefaultConfigFile = "env/config"

	// DefaultEnvKey is the default runtime environment key.
	DefaultEnvKey = "runtime.environment"

	// DefaultHostKey is the default server hostname key.
	DefaultHostKey = "server.host"

	// DefaultPortKey is the default server port key.
	DefaultPortKey = "server.port"

	// DefaultCertPathKey is the default server certificate path key.
	DefaultCertPathKey = "server.certs"

	// DefaultCertProviderKey is the default server certificate provider key.
	DefaultCertProviderKey = "server.certProvider"
)

// CertificateProvider is the common interface for certificate providers.
type CertificateProvider interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
}

// Filter is a func type that modifies a key returned from the envcfg.
type Filter func(*Envcfg, string) string

// Envcfg handles loading configuration variables from system environment
// variables or from an initial configuration file.
type Envcfg struct {
	config *ini.File

	envVarName string
	configFile string

	envKey          string
	hostKey         string
	portKey         string
	certPathKey     string
	certProviderKey string

	tls *tls.Config

	filters map[string]Filter

	sync.Mutex
}

// New creates a new environment configuration loader.
func New(opts ...Option) (*Envcfg, error) {
	var err error

	// default values
	ec := &Envcfg{
		envVarName:      DefaultVarName,
		configFile:      DefaultConfigFile,
		envKey:          DefaultEnvKey,
		hostKey:         DefaultHostKey,
		portKey:         DefaultPortKey,
		certPathKey:     DefaultCertPathKey,
		certProviderKey: DefaultCertProviderKey,
		filters:         make(map[string]Filter),
	}

	// apply options
	for _, o := range opts {
		o(ec)
	}

	// load environment data from $ENV{$envVarName} or from file $configFile
	if envdata := os.Getenv(ec.envVarName); envdata != "" {
		// if the data is supplied in $ENV{$envVarName}, then base64 decode the data
		var data []byte
		data, err = base64.StdEncoding.DecodeString(envdata)
		if err == nil {
			r := bytes.NewReader(data)
			ec.config, err = ini.Load(r)
		}
	} else {
		ec.config, err = ini.LoadFile(ec.configFile)
	}

	// ensure no err
	if err != nil {
		return nil, err
	}

	// set git style config
	ec.config.SectionNameFunc = ini.GitSectionNameFunc
	ec.config.SectionManipFunc = ini.GitSectionManipFunc
	ec.config.ValueManipFunc = func(val string) string {
		val = strings.TrimSpace(val)

		if str, err := strconv.Unquote(val); err == nil {
			val = str
		}

		return val
	}

	return ec, nil
}

// nameRE matches the definition of standard "$SOME_NAME" identifier.
var nameRE = regexp.MustCompile(`(?i)^\$([a-z][a-z0-9_]*)$`)

// GetKey retrieves the value for the key from the environment, or from the
// initial supplied configuration data.
//
// When the initial value read from the config file or the supplied app
// environment variable is in the form of "$NAME||<default>" or
// "$NAME||<default>||<encoding>". Then the value will be read from from the system
// environment variable $NAME. If that value is empty, then the <default> value
// will be returned instead. If the third, optional parameter is
// supplied then the environment variable value (or the default value) will be
// decoded using the appropriate method.
//
// Current supported <encoding> parameters:
//     base64 -- value should be base64 decoded
//     file   -- value should be read from disk
func (ec *Envcfg) GetKey(key string) string {
	val := ec.config.GetKey(key)

	m := strings.Split(val, "||")
	if (len(m) == 2 || len(m) == 3) && nameRE.MatchString(m[0]) {
		// config data has $NAME, so read $ENV{$NAME}
		v := os.Getenv(m[0][1:])

		// if empty value, use the default
		if v == "" {
			val = m[1]
		} else {
			val = v
		}

		if len(m) == 3 {
			switch m[2] {
			case "base64":
				if buf, err := base64.StdEncoding.DecodeString(val); err == nil {
					val = string(buf)
				}

			case "file":
				if buf, err := ioutil.ReadFile(val); err == nil {
					val = string(buf)
				}
			}
		}
	}

	// apply filter
	if f, ok := ec.filters[key]; ok {
		return f(ec, val)
	}

	return val
}

// GetString retrieves the value for key from the environment or the supplied
// configuration data, returning it as a string.
//
// NOTE: alias for GetKey.
func (ec *Envcfg) GetString(key string) string {
	return ec.GetKey(key)
}

// GetBool retrieves the value for key from the environment, or the supplied
// configuration data, returning it as a bool.
func (ec *Envcfg) GetBool(key string) bool {
	b, _ := strconv.ParseBool(ec.GetKey(key))
	return b
}

// GetFloat retrieves the value for key from the environment, or the supplied
// configuration data, returning it as a float64. Uses bitSize as the
// precision.
func (ec *Envcfg) GetFloat(key string, bitSize int) float64 {
	f, _ := strconv.ParseFloat(ec.GetKey(key), bitSize)
	return f
}

// GetInt64 retrieves the value for key from the environment, or the supplied
// configuration data, returning it as a int64. Uses base and bitSize to parse.
func (ec *Envcfg) GetInt64(key string, base, bitSize int) int64 {
	i, _ := strconv.ParseInt(ec.GetKey(key), base, bitSize)
	return i
}

// GetUint64 retrieves the value for key from the environment, or the supplied
// configuration data, returning it as a uint64. Uses base and bitSize to
// parse.
func (ec *Envcfg) GetUint64(key string, base, bitSize int) uint64 {
	u, _ := strconv.ParseUint(ec.GetKey(key), base, bitSize)
	return u
}

// GetInt retrieves the value for key from the environment, or the supplied
// configuration data, returning it as a int. Expects numbers to be base 10 and
// no larger than 32 bits.
func (ec *Envcfg) GetInt(key string) int {
	i, _ := strconv.Atoi(ec.GetKey(key))
	return i
}

// Env retrieves the value for the runtime environment key.
func (ec *Envcfg) Env() string {
	return ec.GetKey(ec.envKey)
}

// Host retrieves the value for the server host key.
func (ec *Envcfg) Host() string {
	return ec.GetKey(ec.hostKey)
}

// Port retrieves the value for the server port key.
func (ec *Envcfg) Port() int {
	return ec.GetInt(ec.portKey)
}

// CertPath retrieves the value for the server certificate path key.
func (ec *Envcfg) CertPath() string {
	return ec.GetKey(ec.certPathKey)
}

// AutocertManager returns an autocert.Manager.
func (ec *Envcfg) AutocertManager() *autocert.Manager {
	// setup letsencrypt autocert manager
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(ec.Host()),
		Cache:      autocert.DirCache(ec.CertPath()),
	}
}

// AutocertDNSManager returns an autocertdns.Manager.
func (ec *Envcfg) AutocertDNSManager(email string, provisioner autocertdns.Provisioner) *autocertdns.Manager {
	return &autocertdns.Manager{
		Prompt:      autocert.AcceptTOS,
		Domain:      ec.Host(),
		Email:       "admin@" + ec.Host(),
		CacheDir:    ec.CertPath(),
		Provisioner: provisioner,
	}
}

// CertProvider returns the configured certificate provider.
func (ec *Envcfg) CertProvider() CertificateProvider {
	provider := ec.GetString(ec.certProviderKey)
	var params []string
	if i := strings.Index(provider, ":"); i != -1 {
		provider, params = strings.TrimSpace(provider[:i]), strings.Split(provider[i+1:], ":")
	}

	if provider == "" || provider == "auto" {
		return ec.AutocertManager()
	}

	if provider != "dns" {
		panic("unknown certificate provider type")
	}

	// "typ:domain:email:token"
	if len(params) < 4 {
		panic("invalid certificate provider params")
	}

	var provisioner autocertdns.Provisioner
	switch params[0] {
	case "godo", "godop", "do", "digitalocean":
		var err error
		provisioner, err = godop.New(
			godop.Domain(params[1]),
			godop.GodoClientToken(context.Background(), params[3]),
		)
		if err != nil {
			panic(err)
		}

	default:
		panic("invalid certificate provisioner type")
	}

	return ec.AutocertDNSManager(params[2], provisioner)
}

// TLS retrieves the TLS configuration, using the provided certificate
// provider.
func (ec *Envcfg) TLS(certProvider CertificateProvider) *tls.Config {
	if certProvider == nil {
		certProvider = ec.CertProvider()
	}

	ec.Lock()
	defer ec.Unlock()

	if ec.tls == nil {
		ec.tls = &tls.Config{
			NextProtos:     []string{"h2", "http/1.1"},
			ServerName:     ec.Host(),
			GetCertificate: certProvider.GetCertificate,

			// qualys A+ settings
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			},
		}
	}

	return ec.tls
}
