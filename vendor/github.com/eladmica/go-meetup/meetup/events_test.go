package meetup

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetEvent(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/:urlname/events/:id"

	mock, err := getMockResponse("event.json")
	if err != nil {
		t.Errorf("unexpected error in getMockResponse: %v", err)
	}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, string(mock))
	})

	actual, err := testClient.GetEvent(":urlname", ":id")
	if err != nil {
		t.Errorf("unexpected error in GetEvent: %v", err)
	}

	expected := &Event{
		Created:       1476304026000,
		ID:            "234811452",
		Name:          "May 2017 NY Tech Meetup and Afterparty",
		RSVPLimit:     400,
		Status:        "upcoming",
		Time:          1494370800000,
		Updated:       1492530555000,
		UTCOffset:     -14400000,
		WaitlistCount: 0,
		YesRSVPCount:  120,
		Link:          "https://www.meetup.com/ny-tech/events/234811452/",
		Description:   "Join us for NYC's most famous and longest running monthly tech event!",
		Visibility:    "public",
		Venue: &EventVenue{
			ID:                   17858792,
			Name:                 "NYU Skirball center ",
			Lat:                  40.729461669921875,
			Lon:                  -73.99783325195312,
			Repinned:             false,
			Address1:             "566 LaGuardia Place at Washington Square",
			City:                 "New York",
			Country:              "us",
			LocalizedCountryName: "USA",
			ZIP:                  "",
			State:                "NY",
		},
		Group: &EventGroup{
			Created:  1096140622000,
			Name:     "NY Tech Meetup",
			ID:       176399,
			JoinMode: "open",
			Lat:      40.720001220703125,
			Lon:      -74,
			URLName:  "ny-tech",
			Who:      "NYC Technologists",
		},
		Fee: &EventFee{
			Accepts:     "wepay",
			Amount:      10,
			Currency:    "USD",
			Description: "per person",
			Label:       "price",
			Required:    true,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("GetEvent: expected: %+v, actual: %+v", expected, actual)
	}
}

func TestGetEvents(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/:urlname/events"
	testParams := &GetEventsParams{
		Desc: true,
		Page: 5,
	}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, fmt.Sprintf("%v?desc=%v&page=%v", testEndpoint, testParams.Desc, testParams.Page))
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.GetEvents(":urlname", testParams)
	if err != nil {
		t.Errorf("unexpected error in GetEvents: %v", err)
	}
}

func TestFindEvents(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/find/events"

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.FindEvents(nil)
	if err != nil {
		t.Errorf("unexpected error in FindEvents: %v", err)
	}
}

func TestGetRecommendedEventsEvents(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/recommended/events"

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.GetRecommendedEvents(nil)
	if err != nil {
		t.Errorf("unexpected error in GetRecommendedEvents: %v", err)
	}
}
