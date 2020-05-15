package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("creating request failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Add our headers.
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))

	// Get the response.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("doing request failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the response.
	var s Response
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		http.Error(w, fmt.Sprintf("decoding json failed: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("response: %#v\n", s)

	// Encode the response and pretty print.
	json, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("encoding json failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Print the response.
	fmt.Fprintf(w, string(json))
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	fmt.Printf("int: %d\n", i)

	t = &Time{time.Unix(i, 0)}
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	fmt.Println(t.Time.String())
	d := time.Until(t.Time)
	if d <= 0 {
		return []byte(`""`), nil
	}

	// Get the duration.
	s := fmt.Sprintf(`"%s"`, strings.ToLower(humanDuration(d)))
	return []byte(s), nil

}

// RateLimit defines the data type for tracking a rate limit.
type RateLimit struct {
	Limit     int64 `json:"limit,omitempty"`
	Remaining int64 `json:"remaining,omitempty"`
	Reset     Time  `json:"reset,omitempty"`
}

// Resources defines the resources data type.
type Resources struct {
	Core                RateLimit `json:"core,omitempty"`
	Search              RateLimit `json:"search,omitempty"`
	GraphQL             RateLimit `json:"graphql,omitempty"`
	IntegrationManifest RateLimit `json:"integration_manifest,omitempty"`
}

// Response defines the response type.
type Response struct {
	Resources Resources `json:"resources,omitempty"`
	Rate      RateLimit `json:"rate,omitempty"`
}

// humanDuration returns a human-readable approximation of a duration
// (eg. "About a minute", "4 hours ago", etc.).
// This comes from: https://github.com/moby/moby/blob/master/vendor/github.com/docker/go-units/duration.go
func humanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds == 1 {
		return "1 second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	} else if hours := int(d.Hours() + 0.5); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("%d hours", hours)
	} else if hours < 24*7*2 {
		return fmt.Sprintf("%d days", hours/24)
	} else if hours < 24*30*2 {
		return fmt.Sprintf("%d weeks", hours/24/7)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("%d months", hours/24/30)
	}
	return fmt.Sprintf("%d years", int(d.Hours())/24/365)
}
