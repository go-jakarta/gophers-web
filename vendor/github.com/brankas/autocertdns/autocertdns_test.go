package autocertdns

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"

	"github.com/brankas/autocertdns/godop"
)

func TestRenew(t *testing.T) {
	ctxt := context.Background()

	doClient, err := godoClient(ctxt)
	if err != nil {
		t.Fatal(err)
	}

	m := &Manager{
		Prompt:      AcceptTOS,
		CacheDir:    "cache",
		Email:       "kenneth.shaw@brank.as",
		Domain:      "aoeu-dev.brank.as",
		Provisioner: godop.New(doClient, "brank.as"),
	}

	err = m.Renew(ctxt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func godoClient(ctxt context.Context) (*godo.Client, error) {
	tok, err := ioutil.ReadFile(".godo-token")
	if err != nil {
		return nil, err
	}

	return godo.NewClient(oauth2.NewClient(
		ctxt,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: strings.TrimSpace(string(tok)),
			},
		),
	)), nil
}
