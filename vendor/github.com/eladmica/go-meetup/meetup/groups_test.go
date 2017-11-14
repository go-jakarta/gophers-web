package meetup

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetGroup(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/:urlname"

	mock, err := getMockResponse("group.json")
	if err != nil {
		t.Errorf("unexpected error in getMockResponse: %v", err)
	}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, string(mock))
	})

	actual, err := testClient.GetGroup(":urlname")
	if err != nil {
		t.Errorf("unexpected error in GetGroup: %v", err)
	}

	expected := &Group{
		ID:                   176399,
		Name:                 "NY Tech Meetup",
		Link:                 "https://www.meetup.com/ny-tech/",
		URLName:              "ny-tech",
		Description:          "<p><strong>Interested in learning more about all of our membership levels or looking for more detailed information on NYTM?</strong><span> Head over to our full website at </span><a href...",
		Created:              1096140622000,
		City:                 "New York",
		Country:              "US",
		LocalizedCountryName: "USA",
		State:                "NY",
		JoinMode:             "open",
		Visibility:           "public",
		Lat:                  40.72,
		Lon:                  -74,
		Members:              53947,
		Organizer: &GroupOrganizer{
			ID:   164510652,
			Name: "NY Tech Meetup",
			Bio:  "NY Tech Meetup is the world's largest Meetup group. We host a monthly gathering to showcase great technology being developed in New York. The Meetup group is operated by NY Tech Alliance, a non-profit organization.",
		},
		Who: "NYC Technologists",
		Photo: &GroupPhoto{
			ID:          450817043,
			HighResLink: "https://secure.meetupstatic.com/photos/event/9/0/b/3/highres_450817043.jpeg",
			PhotoLink:   "https://secure.meetupstatic.com/photos/event/9/0/b/3/600_450817043.jpeg",
			ThumbLink:   "https://secure.meetupstatic.com/photos/event/9/0/b/3/thumb_450817043.jpeg",
			Type:        "event",
			BaseURL:     "https://secure.meetupstatic.com",
		},
		KeyPhoto: &GroupKeyPhoto{
			ID:          449694535,
			HighResLink: "https://secure.meetupstatic.com/photos/event/d/5/0/7/highres_449694535.jpeg",
			PhotoLink:   "https://secure.meetupstatic.com/photos/event/d/5/0/7/600_449694535.jpeg",
			ThumbLink:   "https://secure.meetupstatic.com/photos/event/d/5/0/7/thumb_449694535.jpeg",
			Type:        "event",
			BaseURL:     "https://secure.meetupstatic.com",
		},
		Timezone: "US/Eastern",
		NextEvent: &GroupNextEvent{
			ID:           "234811452",
			Name:         "May 2017 NY Tech Meetup and Afterparty - Creative Tech Theme",
			YesRSVPCount: 120,
			Time:         1494370800000,
			UTCOffset:    -14400000,
		},
		Category: &GroupCategory{
			ID:        34,
			Name:      "Tech",
			ShortName: "Tech",
			SortName:  "Tech",
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("GetGroup: expected: %+v, actual: %+v", expected, actual)
	}
}

func TestGetSimilarGroups(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/:urlname/similar_groups"

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.GetSimilarGroups(":urlname")
	if err != nil {
		t.Errorf("unexpected error in GetSimilarGroups: %v", err)
	}
}

func TestFindGroups(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/find/groups"

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, testEndpoint)
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.FindGroups(nil)
	if err != nil {
		t.Errorf("unexpected error in FindGroups: %v", err)
	}
}
