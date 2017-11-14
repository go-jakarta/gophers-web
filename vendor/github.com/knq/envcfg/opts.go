package envcfg

// Option is an Envcfg option.
type Option func(*Envcfg)

// VarName is an option that sets the name of the environment variable to
// retrieve the configuration data from.
func VarName(name string) Option {
	return func(ec *Envcfg) {
		ec.envVarName = name
	}
}

// ConfigFile is an option that sets the file path to read data from.
func ConfigFile(path string) Option {
	return func(ec *Envcfg) {
		ec.configFile = path
	}
}

// KeyFilter is an option that adds a key filter.
func KeyFilter(key string, f Filter) Option {
	return func(ec *Envcfg) {
		ec.filters[key] = f
	}
}

// EnvKey is an option that sets the runtime environment key.
func EnvKey(key string) Option {
	return func(ec *Envcfg) {
		ec.envKey = key
	}
}

// HostKey is an option that sets the server host key.
func HostKey(key string) Option {
	return func(ec *Envcfg) {
		ec.hostKey = key
	}
}

// PortKey is an option that sets the server port key.
func PortKey(key string) Option {
	return func(ec *Envcfg) {
		ec.portKey = key
	}
}

// CertPathKey is an option that sets the server certificate path key.
func CertPathKey(key string) Option {
	return func(ec *Envcfg) {
		ec.certPathKey = key
	}
}
