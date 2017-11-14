// Package autocertdns provides autocertificate renewal from LetsEncrypt using
// DNS-01 challenges.
package autocertdns

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/knq/pemutil"

	"golang.org/x/crypto/acme"
)

const (
	acmeKey = "acme_account.key"

	LetsEncryptURL        = acme.LetsEncryptURL
	LetsEncryptStagingURL = "https://acme-staging.api.letsencrypt.org/directory"
)

// Provisioner is the shared interface for providers that can provision DNS
// records.
type Provisioner interface {
	// Provision provisions a DNS entry of typ (always TXT), for the FQDN name
	// and with the provided token.
	Provision(ctxt context.Context, typ, name, token string) error

	// Unprovision unprovisions a DNS entry of typ (always TXT), for the FQDN
	// name and with the provided token.
	Unprovision(ctxt context.Context, typ, name, token string) error
}

type Manager struct {
	DirectoryURL string
	Prompt       func(string) bool
	CacheDir     string
	Email        string
	Domain       string
	Provisioner  Provisioner
	Logf         func(string, ...interface{})

	mu sync.Mutex
}

func (m *Manager) log(s string, v ...interface{}) {
	if m.Logf != nil {
		m.Logf(s, v...)
	}
}

func (m *Manager) errf(s string, v ...interface{}) error {
	err := fmt.Errorf(s, v...)
	m.log("ERROR: %v", err)
	return err
}

func (m *Manager) Renew(ctxt context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	if m.Email == "" {
		return m.errf("must provide Email")
	}
	if m.Prompt == nil {
		return m.errf("must provide Prompt")
	}
	if m.Provisioner == nil {
		return m.errf("must provide Provisioner")
	}

	acmePath := m.CacheDir + "/" + acmeKey
	store := pemutil.Store{}

	// try to load cached credentials
	err = store.LoadFile(acmePath)
	if err != nil && os.IsNotExist(err) {
		store, err = pemutil.GenerateECKeySet(elliptic.P256())
		if err != nil {
			return m.errf("could not generate ec key set: %v", err)
		}
		err = os.MkdirAll(m.CacheDir, 0700)
		if err != nil {
			return m.errf("could not create cache directory: %v", err)
		}

		var buf []byte
		buf, err = store.Bytes()
		if err != nil {
			return m.errf("could not generate PEM: %v", err)
		}
		err = ioutil.WriteFile(acmePath, buf, 0600)
		if err != nil {
			return m.errf("could not save PEM: %v", err)
		}
	} else if err != nil {
		return m.errf("unexpected error encountered: %v", err)
	}

	// grab key
	key, ok := store[pemutil.ECPrivateKey].(*ecdsa.PrivateKey)
	if !ok {
		return m.errf("expected ec private key")
	}

	// create acme client
	directoryURL := m.DirectoryURL
	if directoryURL == "" {
		directoryURL = LetsEncryptURL
	}
	client := &acme.Client{
		Key:          key,
		DirectoryURL: directoryURL,
	}

	// register domain
	_, err = client.Register(ctxt, &acme.Account{
		Contact: []string{"mailto:" + m.Email},
	}, m.Prompt)
	if ae, ok := err.(*acme.Error); err == nil || ok && ae.StatusCode == http.StatusConflict {
		// already registered account
	} else if err != nil {
		return m.errf("could not register with ACME provider: %v", err)
	}

	// create authorize challenges
	authz, err := client.Authorize(ctxt, m.Domain)
	if err != nil {
		return m.errf("could not authorize with ACME provider: %v", err)
	}

	// grab dns challenge
	var challenge *acme.Challenge
	for _, c := range authz.Challenges {
		if c.Type == "dns-01" {
			challenge = c
			break
		}
	}
	if challenge == nil {
		return m.errf("no dns-01 challenge was provided by the ACME provider")
	}

	// exchange dns challenge
	tok, err := client.DNS01ChallengeRecord(challenge.Token)
	if err != nil {
		return m.errf("could not generate token for ACME challenge: %v", err)
	}

	// provision TXT under _acme-challenge.<domain>
	err = m.Provisioner.Provision(ctxt, "TXT", "_acme-challenge."+m.Domain, tok)
	if err != nil {
		return m.errf("could not provision dns-01 TXT challenge: %v", err)
	}
	defer m.Provisioner.Unprovision(ctxt, "TXT", "_acme-challenge."+m.Domain, tok)

	// accept challenge
	_, err = client.Accept(ctxt, challenge)
	if err != nil {
		return m.errf("could not accept ACME challenge: %v", err)
	}

	// wait for authorization
	authz, err = client.WaitAuthorization(ctxt, authz.URI)
	if err != nil {
		return err
	} else if authz.Status != acme.StatusValid {
		return m.errf("dns-01 challenge is %v", authz.Status)
	}

	return nil
}

// GetCertificate
func (m *Manager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return nil, nil
}

// AcceptTOS is a util func that always returns true to indicate acceptance of
// the underlying ACME server's Terms of Service during account registration.
func AcceptTOS(string) bool {
	return true
}
