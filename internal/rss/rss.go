package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

var rssTimeFormats = []string{
	time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC822Z,  // "02 Jan 06 15:04 -0700"
	time.RFC822,   // "02 Jan 06 15:04 MST"
	time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
}

// ParseRSSTime tries multiple layouts until one works
func ParseRSSTime(value string) (time.Time, error) {
	var t time.Time
	var err error
	for _, layout := range rssTimeFormats {
		t, err = time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, err
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request to '%v': %v", feedURL, err)
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting response from '%v': %v", feedURL, err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body from '%v': %v", feedURL, err)
	}

	var rss RSSFeed
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling data from %v: %v", feedURL, err)
	}
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)

	for idx := range rss.Channel.Item {
		rss.Channel.Item[idx].Title = html.UnescapeString(rss.Channel.Item[idx].Title)
		rss.Channel.Item[idx].Description = html.UnescapeString(rss.Channel.Item[idx].Description)
	}

	return &rss, nil
}
