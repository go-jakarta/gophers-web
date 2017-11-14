package meetup

import "fmt"

// Group represents a Meetup group
type Group struct {
	ID                   int             `json:"id"`
	Name                 string          `json:"name"`
	Link                 string          `json:"link"`
	URLName              string          `json:"urlname"`
	Description          string          `json:"description"`
	Created              int             `json:"created"`
	City                 string          `json:"city"`
	Country              string          `json:"country"`
	LocalizedCountryName string          `json:"localized_country_name"`
	State                string          `json:"state"`
	JoinMode             string          `json:"join_mode"`
	Visibility           string          `json:"visibility"`
	Lat                  float64         `json:"lat"`
	Lon                  float64         `json:"lon"`
	Members              int             `json:"members"`
	Who                  string          `json:"who"`
	Timezone             string          `json:"timezone"`
	WelcomeMessage       string          `json:"welcome_message"`
	NextEvent            *GroupNextEvent `json:"next_event"`
	Organizer            *GroupOrganizer `json:"organizer"`
	Photo                *GroupPhoto     `json:"group_photo"`
	KeyPhoto             *GroupKeyPhoto  `json:"key_photo"`
	Category             *GroupCategory  `json:"category"`
}

// GroupNextEvent represents the group's next event
type GroupNextEvent struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	YesRSVPCount int    `json:"yes_rsvp_count"`
	Time         int64  `json:"time"`
	UTCOffset    int    `json:"utc_offset"`
}

// GroupOrganizer represents the group's organizer
type GroupOrganizer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Bio   string `json:"bio"`
	Photo struct {
		ID          int    `json:"id"`
		HighResLink string `json:"highres_link"`
		PhotoLink   string `json:"photo_link"`
		ThumbLink   string `json:"thumb_link"`
		Type        string `json:"type"`
		BaseURL     string `json:"base_url"`
	} `json:"photo"`
}

// GroupPhoto represents the group's photo
type GroupPhoto struct {
	ID          int    `json:"id"`
	HighResLink string `json:"highres_link"`
	PhotoLink   string `json:"photo_link"`
	ThumbLink   string `json:"thumb_link"`
	Type        string `json:"type"`
	BaseURL     string `json:"base_url"`
}

// GroupKeyPhoto represents the group's primary photo
type GroupKeyPhoto struct {
	ID          int    `json:"id"`
	HighResLink string `json:"highres_link"`
	PhotoLink   string `json:"photo_link"`
	ThumbLink   string `json:"thumb_link"`
	Type        string `json:"type"`
	BaseURL     string `json:"base_url"`
}

// GroupCategory represents the group's category
type GroupCategory struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortname"`
	SortName  string `json:"sort_name"`
}

// GetGroup gets a group's information using the group's url name
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/#get
func (c *Client) GetGroup(groupURLName string) (*Group, error) {
	url := fmt.Sprintf("%v/%v", c.BaseURL, groupURLName)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var group *Group
	err = c.Do(req, &group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

// GetSimilarGroups gets a listing of similar groups based on a given group's url name
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/similar_groups/
func (c *Client) GetSimilarGroups(groupURLName string) ([]*Group, error) {
	url := fmt.Sprintf("%v/%v/similar_groups", c.BaseURL, groupURLName)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var groups []*Group
	err = c.Do(req, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// FindGroupsParams represents optional parameters for FindGroups
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/groups/
type FindGroupsParams struct {
	Order               string  `url:"order,omitempty"`
	Category            string  `url:"category,omitempty"`
	Country             string  `url:"country,omitempty"`
	FallbackSuggestions bool    `url:"fallback_suggestions,omitempty"`
	Fields              string  `url:"fields,omitempty"`
	Filter              string  `url:"filter,omitempty"`
	Lat                 float64 `url:"lat,omitempty"`
	Lon                 float64 `url:"lon,omitempty"`
	Location            string  `url:"location,omitempty"`
	Radius              int     `url:"radius,omitempty"`
	SelfGroups          string  `url:"self_groups,omitempty"`
	Text                string  `url:"text,omitempty"`
	TopicId             string  `url:"topic_id,omitempty"`
	UpcomingEvents      bool    `url:"upcoming_events,omitempty"`
	ZIP                 string  `url:"zip,omitempty"`
}

// FindGroups gets a listing of group based on the search parameters
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/groups/
func (c *Client) FindGroups(params *FindGroupsParams) ([]*Group, error) {
	url := fmt.Sprintf("%v/find/groups", c.BaseURL)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var groups []*Group
	err = c.Do(req, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
