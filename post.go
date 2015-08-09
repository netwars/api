package main

import (
	"errors"
	"strings"
	"time"

	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	postModificationSelector = "p.post_modified"
)

// Post ...
type Post struct {
	Serial     int64      `json:"serial"`
	TopicID    int64      `json:"topicId"`
	CreatedAt  *time.Time `json:"createdAt"`
	CreatedBy  string     `json:"createdBy"`
	Modified   bool       `json:"modified"`
	ModifiedAt *time.Time `json:"modifiedAt"`
	ModifiedBy string     `json:"modifiedBy"`
	Content    string     `json:"content"`
}

// NewTopicFromDocument parse given document to find matching patterns and returns slice of Post instances if it is possible.
func NewPostsFromDocument(doc *goquery.Document) (posts []*Post, err error) {
	parts := strings.Split(doc.Url.String(), "/")
	if len(parts) < 3 || parts[len(parts)-2] != "temat" {
		return nil, errors.New("malformed topic url")
	}

	topicID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return nil, errors.New("malformed topic id in url")
	}

	doc.Find("div.post[id^='post_']").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var createdAt *time.Time
		var modifiedAt *time.Time
		var serial int64

		serialText := s.Find("span.numerek_posta").Text()
		serial, err = strconv.ParseInt(serialText[2:len(serialText)-1], 10, 64)
		if err != nil {
			return false
		}

		createdAt, err := parseDate(s.Find("div.p2_data").Text())
		if err != nil {
			return false
		}

		post := &Post{
			TopicID:   topicID,
			Serial:    serial,
			Content:   cleanupPostContent(s.Find("div.post_body")).Text(),
			CreatedAt: createdAt,
			CreatedBy: s.Find("div.p2_nick a.nick").Text(),
		}

		if mod := s.Find(postModificationSelector).Text(); mod != "" {
			mod = strings.Replace(mod, "Zmieniony ", "", -1)
			mod = strings.Replace(mod, "przez ", "", -1)
			modParts := strings.Split(mod, " ")

			modifiedAt, err = parseDate(modParts[0] + " " + modParts[1])
			if err != nil {
				return false
			}

			post.Modified = true
			post.ModifiedAt = modifiedAt
			post.ModifiedBy = modParts[2]
		}

		posts = append(posts, post)

		return true
	})
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (p Post) String() string {
	return strconv.FormatInt(p.TopicID, 10) + ":" + strconv.FormatInt(p.Serial, 10) + " - " + p.CreatedAt.String() + " " + p.CreatedBy
}

func cleanupPostContent(s *goquery.Selection) *goquery.Selection {
	s.RemoveFiltered("div.cite")
	s.Find("br").ReplaceWithHtml("\n")

	return s
}
