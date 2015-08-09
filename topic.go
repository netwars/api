package main

import (
	"errors"
	"strings"
	"time"

	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Topic ...
type Topic struct {
	ID        int        `json:"id"`
	ForumID   int        `json:"forumId"`
	Title     string     `json:"title"`
	Posts     []*Post    `json:"posts"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// NewTopicFromDocument parse given document to find matching patterns and returns Topic instance if it is possible.
func NewTopicFromDocument(doc *goquery.Document) (*Topic, error) {
	parts := strings.Split(doc.Url.String(), "/")
	if len(parts) < 3 || parts[len(parts)-2] != "temat" {
		return nil, errors.New("malformed topic url")
	}

	topicID, err := strconv.ParseInt(parts[len(parts)-1], 10, 32)
	if err != nil {
		return nil, errors.New("malformed topic id in url")
	}

	forumLink, _ := doc.Find("ul.forum_navi a[href^='/forum/']").Attr("href")
	if forumLink == "" {
		return nil, errors.New("missing forum link")
	}

	forumID, err := strconv.ParseInt(forumLink[7:], 10, 32)
	if err != nil {
		return nil, errors.New("malformed forum id in url")
	}

	title := doc.Find("title").First().Text()
	if title == "" {
		return nil, errors.New("missing title in document")
	}

	dateRaw := doc.Find(".posthead .p2_data").Last().Text()
	date, err := parseDate(dateRaw)
	if err != nil {
		return nil, err
	}

	return &Topic{
		Title:     title,
		ID:        int(topicID),
		ForumID:   int(forumID),
		UpdatedAt: date,
	}, nil
}
