package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jthughes/gatorcli/internal/database"
)

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

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {

	request, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "gator")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal data")
	}

	// Unescape values from the HTML
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for index, item := range feed.Channel.Item {
		feed.Channel.Item[index].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[index].Description = html.UnescapeString(item.Description)
	}
	return &feed, nil
}

func scrapeFeeds(ctx context.Context, s *state) error {
	oldestFeedUrl := ""
	for {
		feedEntry, err := s.dbq.GetNextFeedToFetch(ctx)
		if err != nil {
			return fmt.Errorf("unable to fetch next feed: %w", err)
		}
		if oldestFeedUrl == feedEntry.Url {
			return nil
		}
		if oldestFeedUrl == "" {
			oldestFeedUrl = feedEntry.Url
		}
		err = s.dbq.MarkFeedFetch(ctx, database.MarkFeedFetchParams{
			LastFetchedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			ID: feedEntry.ID,
		})
		if err != nil {
			return fmt.Errorf("unable to mark fetched feed as fetched: %w", err)
		}
		feed, err := fetchFeed(ctx, feedEntry.Url)
		if err != nil {
			return fmt.Errorf("unable to fetch feed from url: %w", err)
		}
		fmt.Printf("Fetching %s from <%s>\n", feedEntry.Name, feedEntry.Url)
		for _, item := range feed.Channel.Item {
			err = addPost(item, feedEntry, ctx, s)
			if err == nil {
				fmt.Printf("Found post: %s (published '%s')\n", item.Title, item.PubDate)
			}
		}

	}
}

func addPost(post RSSItem, feedEntry database.Feed, ctx context.Context, s *state) error {

	var published_time time.Time

	timeFormats := []string{
		time.RFC1123Z,    //    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
		time.RFC1123,     //     = "Mon, 02 Jan 2006 15:04:05 MST"
		time.Layout,      //      = "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
		time.ANSIC,       //       = "Mon Jan _2 15:04:05 2006"
		time.UnixDate,    //    = "Mon Jan _2 15:04:05 MST 2006"
		time.RubyDate,    //    = "Mon Jan 02 15:04:05 -0700 2006"
		time.RFC822,      //      = "02 Jan 06 15:04 MST"
		time.RFC822Z,     //     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
		time.RFC850,      //      = "Monday, 02-Jan-06 15:04:05 MST"
		time.RFC3339Nano, // = "2006-01-02T15:04:05.999999999Z07:00"
		time.RFC3339,     //     = "2006-01-02T15:04:05Z07:00"

		// Handy time stamps.
		time.StampNano,  //  = "Jan _2 15:04:05.000000000"
		time.StampMicro, // = "Jan _2 15:04:05.000000"
		time.StampMilli, // = "Jan _2 15:04:05.000"
		time.Stamp,      //      = "Jan _2 15:04:05"
		time.DateTime,   //   = "2006-01-02 15:04:05"
		time.DateOnly,   //   = "2006-01-02"
		time.TimeOnly,   //   = "15:04:05"
		time.Kitchen,    //     = "3:04PM"
	}

	for _, format := range timeFormats {
		parsedTime, err := time.Parse(format, post.PubDate)
		if err == nil {
			published_time = parsedTime
			break
		}
	}

	_, err := s.dbq.CreatePost(ctx, database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       post.Title,
		Url:         post.Link,
		Description: post.Description,
		PublishedAt: sql.NullTime{
			Time:  published_time,
			Valid: true,
		},
		FeedID: feedEntry.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to add new post to database: %w", err)
	}
	return nil
}
