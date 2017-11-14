package meetup

import (
	"testing"
)

func TestAuthenticateRequestUsingKeyAuth(t *testing.T) {
	setup()
	defer teardown()

	key := "secret"
	testClient.Authentication = NewKeyAuth(key)

	req, err := testClient.NewRequest("GET", testClient.BaseURL, nil)
	if err != nil {
		t.Errorf("unexpected error in NewRequest: %v", err)
	}

	err = testClient.Authentication.AuthenticateRequest(req)
	if err != nil {
		t.Errorf("unexpected error in AuthenticateRequest: %v", err)
	}

	if actual := req.URL.Query().Get("key"); actual != key {
		t.Errorf("expected key: %v, actual: %v", key, actual)
	}
}
