package meetup

import "fmt"

// Event represents a Meetup event
type Event struct {
	Created       int         `json:"created"`
	Duration      int         `json:"duration"`
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Status        string      `json:"status"`
	Time          int         `json:"time"`
	Updated       int         `json:"updated"`
	UTCOffset     int         `json:"utc_offset"`
	WaitlistCount int         `json:"waitlist_count"`
	RSVPLimit     int         `json:"rsvp_limit"`
	YesRSVPCount  int         `json:"yes_rsvp_count"`
	Link          string      `json:"link"`
	Description   string      `json:"description"`
	Visibility    string      `json:"visibility"`
	Venue         *EventVenue `json:"venue"`
	Group         *EventGroup `json:"group"`
	Fee           *EventFee   `json:"fee"`
}

// EventVenue represents an event venue
type EventVenue struct {
	ID                   int     `json:"id"`
	Name                 string  `json:"name"`
	Lat                  float64 `json:"lat"`
	Lon                  float64 `json:"lon"`
	Repinned             bool    `json:"repinned"`
	Address1             string  `json:"address_1"`
	Address2             string  `json:"address_2"`
	Address3             string  `json:"address_3"`
	City                 string  `json:"city"`
	Country              string  `json:"country"`
	LocalizedCountryName string  `json:"localized_country_name"`
	ZIP                  string  `json:"zip"`
	State                string  `json:"state"`
}

// EventGroup represents an event hosting group
type EventGroup struct {
	Created  int     `json:"created"`
	Name     string  `json:"name"`
	ID       int     `json:"id"`
	JoinMode string  `json:"join_mode"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	URLName  string  `json:"urlname"`
	Who      string  `json:"who"`
}

// EventFee represents an event fee
type EventFee struct {
	Accepts     string  `json:"accepts"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Label       string  `json:"label"`
	Required    bool    `json:"required"`
}

// GetEvents gets a single event using a group's url name and an event's id.
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/events/:id/#get
func (c *Client) GetEvent(groupURLName, eventID string) (*Event, error) {
	url := fmt.Sprintf("%v/%v/events/%v", c.BaseURL, groupURLName, eventID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var event *Event
	err = c.Do(req, &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetEventsParams represents optional parameters for GetEvents
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/events/#list
type GetEventsParams struct {
	Desc   bool   `url:"desc,omitempty"`
	Fields string `url:"fields,omitempty"`
	Page   int    `url:"page,omitempty"`
	Status string `url:"status,omitempty"`
	Scroll string `url:"scroll,omitempty"`
}

// GetEvents gets a listing of all events under the given group's url name
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/events/#list
func (c *Client) GetEvents(groupURLName string, params *GetEventsParams) ([]*Event, error) {
	url := fmt.Sprintf("%v/%v/events", c.BaseURL, groupURLName)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var events []*Event
	err = c.Do(req, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// FindEventsParams represents optional parameters for FindEvents
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/events/
type FindEventsParams struct {
	Lat    float64 `url:"lat,omitempty"`
	Lon    float64 `url:"lon,omitempty"`
	Radius int     `url:"radius,omitempty"`
	Text   string  `url:"text,omitempty"`
	Fields string  `url:"fields,omitempty"`
}

// FindEvents gets a listing of upcoming events based on location
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/events/
func (c *Client) FindEvents(params *FindEventsParams) ([]*Event, error) {
	url := fmt.Sprintf("%v/find/events", c.BaseURL)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var events []*Event
	err = c.Do(req, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetRecommendedEventsParams represents optional parameters for GetRecommendedEvents
// Meetup docs: https://www.meetup.com/meetup_api/docs/recommended/events/
type GetRecommendedEventsParams struct {
	Lat           float64 `json:"lat,omitempty"`
	Lon           float64 `json:"lon,omitempty"`
	Page          int     `json:"page,omitempty"`
	Fields        string  `json:"fields,omitempty"`
	SelfGroups    string  `json:"self_groups,omitempty"`
	TopicCategory int     `json:"topic_category,omitempty"`
}

// GetRecommendedEvents gets a listing of upcoming recommended events
// Meetup docs: https://www.meetup.com/meetup_api/docs/recommended/events/
func (c *Client) GetRecommendedEvents(params *GetRecommendedEventsParams) ([]*Event, error) {
	url := fmt.Sprintf("%v/recommended/events", c.BaseURL)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var events []*Event
	err = c.Do(req, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}
