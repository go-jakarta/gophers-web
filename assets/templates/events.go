package templates

import (
	"time"
)

type Event struct {
	Created       time.Time     `json:"created"`
	Duration      time.Duration `json:"duration"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Status        string        `json:"status"`
	Time          time.Time     `json:"time"`
	Updated       time.Time     `json:"updated"`
	UTCOffset     int           `json:"utc_offset"`
	WaitlistCount int           `json:"waitlist_count"`
	RSVPLimit     int           `json:"rsvp_limit"`
	YesRSVPCount  int           `json:"yes_rsvp_count"`
	Link          string        `json:"link"`
	Description   string        `json:"description"`
	Visibility    string        `json:"visibility"`
	Venue         *Venue        `json:"venue"`
	Group         *Group        `json:"group"`
	Fee           *Fee          `json:"fee"`
}

type Fee struct {
	Accepts     string  `json:"accepts"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Label       string  `json:"label"`
	Required    bool    `json:"required"`
}

type Group struct {
	Created  time.Time `json:"created"`
	Name     string    `json:"name"`
	ID       int       `json:"id"`
	JoinMode string    `json:"join_mode"`
	Lat      float64   `json:"lat"`
	Lon      float64   `json:"lon"`
	URLName  string    `json:"urlname"`
	Who      string    `json:"who"`
}

type Venue struct {
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
