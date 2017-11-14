package meetup

import "net/http"

// Authenticator is the interface that wraps the AuthenticateRequest method.
type Authenticator interface {
	AuthenticateRequest(*http.Request) error
}

// KeyAuth authenticates requests using a user's private key.
// Meetup docs: https://www.meetup.com/meetup_api/auth/#keys
type KeyAuth struct {
	Key string
}

func NewKeyAuth(key string) Authenticator {
	return &KeyAuth{Key: key}
}

// AuthenticateRequest implements the Authenticator interface.
func (auth *KeyAuth) AuthenticateRequest(req *http.Request) error {
	params := req.URL.Query()
	params.Set("key", auth.Key)
	req.URL.RawQuery = params.Encode()
	return nil
}
