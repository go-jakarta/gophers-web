package meetup

import "fmt"

// Topic represents a Meetup topic
type Topic struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URLKey      string `json:"urlkey"`
	GroupCount  int    `json:"group_count"`
	MemberCount int    `json:"member_count"`
	Description string `json:"description"`
	Lang        string `json:"lang"`
}

// TopicCategory represents a high level topic category
type TopicCategory struct {
	ID        int    `json:"id"`
	ShortName string `json:"shortname"`
	Name      string `json:"name"`
	SortName  string `json:"sort_name"`
	Photo     struct {
		ID          int    `json:"id"`
		HighresLink string `json:"highres_link"`
		PhotoLink   string `json:"photo_link"`
		ThumbLink   string `json:"thumb_link"`
		Type        string `json:"type"`
		BaseURL     string `json:"base_url"`
	} `json:"photo"`
	CategoryIds []int `json:"category_ids"`
}

// FindTopics gets a listing of topics based on the query search parameter
// Meetup docs: https://www.meetup.com/meetup_api/docs/:urlname/events/:id/#get
func (c *Client) FindTopics(query string) ([]*Topic, error) {
	url := fmt.Sprintf("%v/find/topics?query=%v", c.BaseURL, query)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var topics []*Topic
	err = c.Do(req, &topics)
	if err != nil {
		return nil, err
	}

	return topics, nil
}

// FindTopicCategoriesParams represents optional parameters for FindTopicCategories
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/topic_categories/
type FindTopicCategoriesParams struct {
	Fields string  `url:"fields,omitempty"`
	Lat    float64 `url:"lat,omitempty"`
	Lon    float64 `url:"lon,omitempty"`
	Radius int     `url:"radius,omitempty"`
}

//  FindTopicCategories gets a listing of high level topic categories
// Meetup docs: https://www.meetup.com/meetup_api/docs/find/topic_categories/
func (c *Client) FindTopicCategories(params *FindTopicCategoriesParams) ([]*TopicCategory, error) {
	url := fmt.Sprintf("%v/find/topic_categories", c.BaseURL)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var topicCategories []*TopicCategory
	err = c.Do(req, &topicCategories)
	if err != nil {
		return nil, err
	}

	return topicCategories, nil
}

// GetRecommendedGroupTopicsParams represents optional parameters for GetRecommendedGroupTopics
// Meetup docs: https://www.meetup.com/meetup_api/docs/recommended/group_topics/
type GetRecommendedGroupTopicsParams struct {
	ExcludeTopics string `url:"exclude_topics,omitempty"`
	Lang          string `url:"lang,omitempty"`
	OtherTopics   string `url:"other_topics,omitempty"`
	Page          int    `url:"page,omitempty"`
	Text          string `url:"text,omitempty"`
}

// GetRecommendedGroupTopics gets a listing of recommended group topics based on a text search or other topics
// Meetup docs: https://www.meetup.com/meetup_api/docs/recommended/group_topics/
func (c *Client) GetRecommendedGroupTopics(params *GetRecommendedGroupTopicsParams) ([]*Topic, error) {
	url := fmt.Sprintf("%v/recommended/group_topics", c.BaseURL)

	url, err := addQueryParams(url, params)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var topics []*Topic
	err = c.Do(req, &topics)
	if err != nil {
		return nil, err
	}

	return topics, nil
}
