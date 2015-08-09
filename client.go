package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Client ...
type Client interface {
	FetchTopic(int) (*Topic, error)
	FetchTopics(int, chan<- *Topic) error
}

type client struct {
	url *url.URL
}

// NewClient ...
func NewClient(u *url.URL) Client {
	return &client{
		url: u,
	}
}

// FetchDocument ...
func (c *client) FetchDocument(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromResponse(resp)
}

// FetchTopic ...
func (c *client) FetchTopic(topicID int) (*Topic, error) {
	doc, err := c.FetchDocument(c.url.String() + "/temat/" + strconv.FormatInt(int64(topicID), 10))
	if err != nil {
		return nil, err
	}

	topic, err := NewTopicFromDocument(doc)
	if err != nil {
		return nil, err
	}

	topic.Posts, err = NewPostsFromDocument(doc)
	if err != nil {
		return nil, err
	}

	return topic, nil
}

// FetchTopics ...
func (c *client) FetchTopics(nbOfPages int, result chan<- *Topic) error {
	ids := []int{ForumIDStarCraft, ForumIDStarCraft2, ForumIDOtherGames, ForumIDOffTopic}

	for _, id := range ids {
		if err := c.FetchTopicsForForum(id, nbOfPages, result); err != nil {
			return err
		}
	}

	return nil
}

// FetchTopicsForForum ...
func (c *client) FetchTopicsForForum(id, nbOfPages int, result chan<- *Topic) error {
	forumURL := c.url.String() + "/forum/" + strconv.FormatInt(int64(id), 10)

	doc, err := c.FetchDocument(forumURL)
	if err != nil {
		return err
	}

	navi := doc.Find("ul.forum_navi a[href^='/forum/']")

	forumLink, _ := navi.Attr("href")
	if forumLink == "" {
		return errors.New("missing forum link")
	}

	lastPageLink := doc.Find("ul.pagination_list li:last-of-type a").Text()
	if lastPageLink == "" {
		return errors.New("missing last page link")
	}

	lastPageID, err := strconv.ParseInt(lastPageLink, 10, 32)
	if err != nil {
		return errors.New("malformed last page ID")
	}

	if int(lastPageID) < nbOfPages {
		nbOfPages = int(lastPageID)
	}
	for pageID := 0; pageID < nbOfPages; pageID++ {
		doc, err := c.FetchDocument(forumURL + "/" + strconv.FormatInt(int64(pageID), 10))
		if err != nil {
			return err
		}

		topicRowsSelection := doc.Find("table td.topic a[href^='/temat/']")

		for i := 0; i < topicRowsSelection.Length(); i++ {
			href, exists := topicRowsSelection.Eq(i).Attr("href")
			if !exists {
				continue
			}

			topicID, err := strconv.ParseInt(href[7:], 10, 32)
			if err != nil {
				return err
			}

			topic, err := c.FetchTopic(int(topicID))
			if err != nil {
				return err
			}

			result <- topic
		}
	}

	return nil
}
