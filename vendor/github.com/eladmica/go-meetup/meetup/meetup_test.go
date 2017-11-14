package meetup

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

const (
	mocksLocation = "./mocks"
)

var (
	// testMux is the HTTP request multiplexer used with the test server.
	testMux *http.ServeMux

	// testClient is the Meetup client being tested.
	testClient *Client

	// testServer is a test HTTP server used to provide mock API responses from Meetup.
	testServer *httptest.Server
)

// setup sets the test server and Meetup client.
func setup() {
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)
	testClient = NewClient(nil)
	testClient.BaseURL = testServer.URL
}

// teardown closes the test HTTP server.
func teardown() {
	testServer.Close()
}

// getMockResponse reads mock Meetup API responses from a file.
func getMockResponse(file string) ([]byte, error) {
	fileName := fmt.Sprintf("%v/%v", mocksLocation, file)

	res, err := ioutil.ReadFile(fileName)
	if err != nil {
		return res, err
	}

	return res, nil
}

func testRequestMethod(t *testing.T, req *http.Request, expected string) {
	if actual := req.Method; actual != expected {
		t.Errorf("expected method: %v, actual %v", expected, actual)
	}
}

func testRequestURL(t *testing.T, req *http.Request, expected string) {
	expectedURL, err := url.Parse(expected)
	if err != nil {
		t.Errorf("unexpected error: badly formatted expected URL: %v", err)
	}
	if actual := req.URL; !reflect.DeepEqual(actual, expectedURL) {
		t.Errorf("expected URL: %+v, actual: %+v", expected, actual)
	}
}

func TestParseRate(t *testing.T) {
	expected := &Rate{
		Limit:     30,
		Remaining: 20,
		Reset:     5,
	}

	header := http.Header{}
	header.Set(headerRateLimit, strconv.Itoa(expected.Limit))
	header.Set(headerRateRemaining, strconv.Itoa(expected.Remaining))
	header.Set(headerRateReset, strconv.Itoa(expected.Reset))

	actual := parseRate(header)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("parseRate(%v): expected: %v, actual %v",
			header, expected, actual)
	}
}

func TestAddQueryParams(t *testing.T) {
	rawURL := testClient.BaseURL

	var testCases = []struct {
		params   interface{}
		expected string
	}{
		{
			params:   nil,
			expected: rawURL,
		},
		{
			params:   struct{}{},
			expected: rawURL,
		},
		{
			params: struct {
				Text string
			}{
				"dummy",
			},
			expected: rawURL,
		},
		{
			params: &struct {
				Text string `url:"text,omitempty"`
			}{
				"",
			},
			expected: rawURL,
		},
		{
			params: &struct {
				Text *string `url:"text"`
			}{
				nil,
			},
			expected: rawURL,
		},
		{
			params: &struct {
				Text string `url:"text"`
			}{
				"dummy",
			},
			expected: fmt.Sprintf("%v?text=%v", rawURL, "dummy"),
		},
		{
			params: &struct {
				Text   string `url:"text,omitempty"`
				Number int    `url:"number,omitempty"`
				Flag   bool   `url:"flag,omitempty"`
			}{
				"dummy",
				5,
				true,
			},
			expected: fmt.Sprintf("%v?flag=%v&number=%v&text=%v", rawURL, true, 5, "dummy"),
		},
	}

	for _, tc := range testCases {
		actual, err := addQueryParams(rawURL, tc.params)
		if err != nil {
			t.Errorf("unexpected error in addQueryParams(%v, %v): %v",
				rawURL, tc.params, err)
		}
		if actual != tc.expected {
			t.Errorf("addQueryParams(%v, %v): expected %v, actual %v",
				rawURL, tc.params, tc.expected, actual)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	var Ptr *struct{} = nil
	var Int int = 0
	var Bool bool = false
	var String string = ""
	var Uint uint = 0
	var Float float32 = 0
	var Map map[struct{}]struct{}
	var Complex complex64 = 0

	var testCases = []struct {
		input    reflect.Value
		expected bool
	}{
		{reflect.ValueOf(Ptr), true},
		{reflect.ValueOf(Int), true},
		{reflect.ValueOf(Bool), true},
		{reflect.ValueOf(String), true},
		{reflect.ValueOf(Uint), true},
		{reflect.ValueOf(Float), true},
		{reflect.ValueOf(Map), true},
		{reflect.ValueOf(5), false},
		{reflect.ValueOf("text"), false},
		{reflect.ValueOf(&Int), false},
		{reflect.ValueOf(true), false},
		{reflect.ValueOf(0.5), false},
		{reflect.ValueOf(Complex), false},
	}

	for _, tc := range testCases {
		actual := isEmpty(tc.input)
		if actual != tc.expected {
			t.Errorf("isEmpty(%v): expected %v, actual %v",
				tc.input, tc.expected, actual)
		}
	}
}
