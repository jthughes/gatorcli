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
	feedEntry, err := s.dbq.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch next feed: %w", err)
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
	for _, item := range feed.Channel.Item {
		fmt.Println("Found post: %s\n", item.Title)
	}
	return nil
}
