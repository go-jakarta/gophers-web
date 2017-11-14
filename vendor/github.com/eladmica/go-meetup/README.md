# go-meetup

go-meetup is a Go client library for the [Meetup API](https://www.meetup.com/meetup_api/)

## Installation
    go get github.com/eladmica/go-meetup/meetup
    
## Example Usage
```
package main

import (
	"fmt"

	"github.com/eladmica/go-meetup/meetup"
)

func main() {
	// Create a client with API Key authentication
	client := meetup.NewClient(nil)
	client.Authentication = meetup.NewKeyAuth("SECRET_KEY")

	// Get events hosted by the NY Tech Meetup group:
	events, err := client.GetEvents("ny-tech", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("NY Tech Meetup group events:")
	for _, event := range events {
		fmt.Printf("Event name: %v\n", event.Name)
	}

	// Find newest groups related to "Web Development":
	params := meetup.FindGroupsParams{
		Order: "newest",
		Text:  "Web Development",
	}

	groups, err := client.FindGroups(&params)
	if err != nil {
		panic(err)
	}

	fmt.Println("Newest Web Development groups:")
	for _, group := range groups[:10] {
		fmt.Printf("Group name: %v. Link: %v\n", group.Name, group.Link)
	}
}
```

## Authentication
The easiest way to authenticate requests is using Meetup [API Key authentication](https://www.meetup.com/meetup_api/auth/#keys).
```
client := meetup.NewClient(nil)
client.Authentication = meetup.NewKeyAuth("SECRET_KEY")
```

Alternatively, the API client accepts any `http.Client` capable of making user authentication. See [Go OAuth2](https://github.com/golang/oauth2/) for information on how to set up such clients.

## Credits
* [go-github](https://github.com/google/go-github) - A very well designed library which influenced the writing of this library.
