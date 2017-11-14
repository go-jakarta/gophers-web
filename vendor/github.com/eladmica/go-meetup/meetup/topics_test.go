package meetup

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestFindTopics(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/find/topics"
	query := "code"

	mock, err := getMockResponse("topics.json")
	if err != nil {
		t.Errorf("unexpected error in getMockResponse: %v", err)
	}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, fmt.Sprintf("%v?query=%v", testEndpoint, query))
		fmt.Fprint(rw, string(mock))
	})

	actual, err := testClient.FindTopics(query)
	if err != nil {
		t.Errorf("unexpected error in FindTopics: %v", err)
	}

	expected := []*Topic{
		{
			ID:          90286,
			Name:        "Social Coding",
			URLKey:      "social-coding",
			GroupCount:  294,
			MemberCount: 110349,
			Description: "Find out what's happening in Social Coding Meetup groups around the world and start meeting up with the ones near you.",
			Lang:        "en_US",
		},
		{
			ID:          44602,
			Name:        "Coding Dojos",
			URLKey:      "coding-dojos",
			GroupCount:  134,
			MemberCount: 42762,
			Description: "Find out what's happening in Coding Dojos Meetup groups around the world and start meeting up with the ones near you.",
			Lang:        "en_US",
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("FindTopics: expected: %+v, actual: %+v", expected, actual)
	}
}

func TestFindTopicCategories(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/find/topic_categories"
	testParams := &FindTopicCategoriesParams{
		Lat: 49.2606,
		Lon: 123.2460,
	}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, fmt.Sprintf("%v?lat=%v&lon=%v", testEndpoint, testParams.Lat, testParams.Lon))
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.FindTopicCategories(testParams)
	if err != nil {
		t.Errorf("unexpected error in FindTopicCategories: %v", err)
	}
}

func TestGetRecommendedGroupTopics(t *testing.T) {
	setup()
	defer teardown()

	testEndpoint := "/recommended/group_topics"
	testParams := &GetRecommendedGroupTopicsParams{Page: 20}

	testMux.HandleFunc(testEndpoint, func(rw http.ResponseWriter, req *http.Request) {
		testRequestMethod(t, req, "GET")
		testRequestURL(t, req, fmt.Sprintf("%v?page=%v", testEndpoint, testParams.Page))
		fmt.Fprint(rw, "[]")
	})

	_, err := testClient.GetRecommendedGroupTopics(testParams)
	if err != nil {
		t.Errorf("unexpected error in GetRecommendedGroupTopics: %v", err)
	}
}
