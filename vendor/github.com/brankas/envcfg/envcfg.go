// Package envcfg provides a common way to load configuration variables from
// the system environment or from based on initial configuration values stored
// on disk or in a base64 encoded environment variable.
package envcfg

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/knq/ini"
	"github.com/yookoala/realpath"
	"golang.org/x/crypto/acme/autocert"

	"github.com/brankas/autocertdns"
	"github.com/brankas/autocertdns/gcdnsp"
	"github.com/brankas/autocertdns/godop"
)

const (
	// DefaultVarName is the default environment variable name to load the
	// initial configuration data from.
	DefaultVarName = "APP_CONFIG"

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

	// DefaultCertWaitKey is the default server certificate wait key.
	DefaultCertWaitKey = "server.certWait"

	// DefaultCertDelayKey is the default server certificate delay key.
	DefaultCertDelayKey = "server.certDelay"

	// DefaultConfigFile is the default file path to load the initial
	// configuration data from.
	DefaultConfigFile = "env/config"

	// DefaultCertPath is the default certificate caching path.
	DefaultCertPath = "env/certs"

	// DefaultEnvironment is the default environment name.
	DefaultEnvironment = "development"

	// DefaultCertWait is the default certificate provider wait duration.
	DefaultCertWait = 180 * time.Second

	// DefaultCertDelay is the default certificate provider propagation delay.
	DefaultCertDelay = 30 * time.Second
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
	certWaitKey     string
	certDelayKey    string

	tls          *tls.Config
	certProvider CertificateProvider

	filters map[string]Filter

	logf func(string, ...interface{})
	errf func(string, ...interface{})

	once sync.Once
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
		certWaitKey:     DefaultCertWaitKey,
		certDelayKey:    DefaultCertDelayKey,
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
			ec.config, err = ini.Load(bytes.NewReader(data))
		}
	} else {
		ec.configFile, err = realpath.Realpath(ec.configFile)
		if err == nil {
			ec.config, err = ini.LoadFile(ec.configFile)
		}
	}
	// ensure no err
	if err != nil {
		return nil, err
	}

	// ensure log funcs always are set
	if ec.logf == nil {
		ec.logf = func(string, ...interface{}) {}
	}
	if ec.errf == nil {
		ec.errf = func(s string, v ...interface{}) {
			ec.logf("ERROR: "+s, v...)
		}
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
//     base64  -- value should be base64 decoded
//     file    -- value should be read from disk
//     relfile -- value should be read from a path on disk relative to the loaded file
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

			case "relfile":
				buf, err := ioutil.ReadFile(filepath.Join(filepath.Dir(ec.configFile), val))
				if err == nil {
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

// GetDuration retrievses the value for key from the environment, or the
// supplied configuration data, returning it as a time.Duration.
func (ec *Envcfg) GetDuration(key string) time.Duration {
	d, _ := time.ParseDuration(ec.GetKey(key))
	return d
}

// MustKey returns the value for key (same as GetKey) but panics if the
// string is empty.
func (ec *Envcfg) MustKey(key string) string {
	val := ec.GetKey(key)
	if val == "" {
		panic(fmt.Sprintf("key %s must be defined", key))
	}
	return val
}

// MustInt returns the value for key (same as GetInt) but panics if the key was
// blank or if the value is an invalid int.
func (ec *Envcfg) MustInt(key string) int {
	val := ec.MustKey(key)
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("key %s must be a valid int", key))
	}
	return i
}

// Env retrieves the value for the runtime environment key.
func (ec *Envcfg) Env() string {
	if env := ec.GetKey(ec.envKey); env != "" {
		return env
	}
	return DefaultEnvironment
}

// Host retrieves the value for the server host key.
func (ec *Envcfg) Host() string {
	return ec.MustKey(ec.hostKey)
}

// Port retrieves the value for the server port key.
func (ec *Envcfg) Port() int {
	return ec.MustInt(ec.portKey)
}

// PortString retrieves the value for the server port key as a string.
func (ec *Envcfg) PortString() string {
	return strconv.Itoa(ec.Port())
}

// CertPath retrieves the value for the server certificate path key.
func (ec *Envcfg) CertPath() string {
	path := ec.GetKey(ec.certPathKey)
	if path == "" {
		path = DefaultCertPath
	}

	// ensure the directory exists
	fi, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		if err := os.MkdirAll(path, 0700); err != nil {
			panic(fmt.Sprintf("cannot create directory: %v", err))
		}
	case err != nil:
		panic(err)
	case !fi.IsDir():
		panic(fmt.Sprintf("%s must be a directory", path))
	}

	return path
}

// CertWait retrieves the certificate provider wait duration.
func (ec *Envcfg) CertWait() time.Duration {
	if d := ec.GetDuration(ec.certWaitKey); d != 0 {
		return d
	}
	return DefaultCertWait
}

// CertDelay retrieves the certificate provider delay duration.
func (ec *Envcfg) CertDelay() time.Duration {
	if d := ec.GetDuration(ec.certDelayKey); d != 0 {
		return d
	}
	return DefaultCertDelay
}

// TLS retrieves the TLS configuration for use by a server.
func (ec *Envcfg) TLS() *tls.Config {
	ec.once.Do(func() {
		// build cert provider
		certProvider := ec.buildCertProvider()
		if certProvider == nil {
			return
		}

		// set tls config
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
	})
	return ec.tls
}

// buildCertProvider builds the CertificateProvider.
func (ec *Envcfg) buildCertProvider() CertificateProvider {
	provider := ec.GetKey(ec.certProviderKey)
	var params []string
	if i := strings.Index(provider, ":"); i != -1 {
		provider, params = strings.TrimSpace(provider[:i]), strings.Split(provider[i+1:], ":")
	}

	switch provider {
	case "", "auto":
		return &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(ec.Host()),
			Cache:      autocert.DirCache(ec.CertPath()),
		}

	case "dns":
		return ec.dnsCertProvider(params)

	case "disk":
		return ec.diskCertProvider(params)

	case "none":
		return nil

	default:
		panic("unknown certificate provider type: " + provider)
	}
}

// dnsCertProvider handles dns:... configs.
//
// General form is <type>:<domain>:<email address>:[param:[param]...]
func (ec *Envcfg) dnsCertProvider(params []string) CertificateProvider {
	if len(params) < 3 {
		panic("invalid dns certificate provider params")
	}

	var provisioner autocertdns.Provisioner
	switch params[0] {

	// godo:mydomain.com:user@mydomain.com:abaoeusntahoustnhaou
	case "godo", "godop", "do", "digitalocean":
		if len(params) != 4 {
			panic("invalid digitalocean dns certificate provider params")
		}

		var err error
		provisioner, err = godop.New(
			godop.Domain(params[1]),
			godop.GodoClientToken(context.Background(), params[3]),
			godop.Logf(ec.logf),
		)
		if err != nil {
			panic(err)
		}

	// clouddns:mydomain.com:user@mydomain.com:managed-zone-name:/path/to/credentials.json
	case "google", "gcdns", "gcdnsp", "gc", "googlecloud", "clouddns", "googleclouddns":
		if len(params) != 5 {
			panic("invalid google cloud dns certificate provider params")
		}

		var err error
		provisioner, err = gcdnsp.New(
			gcdnsp.Domain(params[1]),
			gcdnsp.ManagedZone(params[3]),
			gcdnsp.GoogleServiceAccountCredentialsFile(params[4]),
			gcdnsp.PropagationWait(ec.CertWait()),
			gcdnsp.ProvisionDelay(ec.CertDelay()),
			gcdnsp.Logf(ec.logf),
			gcdnsp.Errorf(ec.errf),
		)
		if err != nil {
			panic(err)
		}

	default:
		panic("invalid certificate provisioner type")
	}

	a := &autocertdns.Manager{
		Prompt:      autocert.AcceptTOS,
		Domain:      ec.Host(),
		Email:       params[2],
		CacheDir:    ec.CertPath(),
		Provisioner: provisioner,
		Logf:        ec.logf,
	}

	if err := a.Run(context.Background()); err != nil {
		panic(fmt.Sprintf("could not provision: %v", err))
	}

	return a
}

// diskCertProvider creates a certificate provider from a certificate and key
// pair stored on disk.
//
// Uses fsnotify to watch for changes to reload immediately.
func (ec *Envcfg) diskCertProvider(params []string) CertificateProvider {
	// "certname:keyname"
	if len(params) < 2 {
		panic("invalid certificate provider params")
	}

	dir := ec.CertPath()
	certPath, keyPath := filepath.Join(dir, params[0]), filepath.Join(dir, params[1])
	dcp, err := newDiskCertProvider(certPath, keyPath, ec.logf, ec.errf)
	if err != nil {
		panic(fmt.Sprintf("unable to load certificate and key from disk: %v", err))
	}

	return dcp
}

// diskCertProvider provides a certificate provider that watches a local file.
type diskCertProvider struct {
	config *Envcfg

	cert *tls.Certificate

	certPath string
	keyPath  string

	sync.RWMutex
}

// newDiskCertProvider creates a disk cert provider that watches dirPath for
// cert and key.
func newDiskCertProvider(certPath, keyPath string, logf, errf func(string, ...interface{})) (*diskCertProvider, error) {
	dcp := &diskCertProvider{
		certPath: certPath,
		keyPath:  keyPath,
	}
	if err := dcp.loadCertAndKey(); err != nil {
		return nil, err
	}

	// create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	// don't close, as it runs forever
	// defer watcher.Close()

	addWatch := func(path string) error {
		if err := watcher.Add(certPath); err != nil {
			return err
		}

		// if this is a valid symlink, watch its target too
		if path, err := filepath.EvalSymlinks(certPath); err == nil && path != certPath {
			if err := watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	}

	go func() {
		for {
			select {
			case <-watcher.Events:
				if err := dcp.loadCertAndKey(); err != nil {
					errf("could not load cert and key: %v", err)
				} else {
					logf("loaded updated certificate")
				}

				// in case we're dealing with symlinks and the target changed,
				// make sure we continue to watch properly
				if err := addWatch(certPath); err != nil {
					errf("could not add watch: %v", err)
				}

			case err := <-watcher.Errors:
				errf("%v", err)
			}
		}
	}()

	// watch the first certificate file
	if err := addWatch(certPath); err != nil {
		return nil, err
	}

	return dcp, nil
}

// loadCertAndKey tries to load the cert and key from disk.
func (dcp *diskCertProvider) loadCertAndKey() error {
	cert, err := tls.LoadX509KeyPair(dcp.certPath, dcp.keyPath)
	if err != nil {
		return err
	}

	dcp.Lock()
	defer dcp.Unlock()

	dcp.cert = &cert
	return nil
}

// GetCertificate satisfies the CertificateProvider interface.
func (dcp *diskCertProvider) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	dcp.RLock()
	defer dcp.RUnlock()

	return dcp.cert, nil
}
