package main

import (
	"net/http"

	"github.com/piotrkowalczuk/rest"
	"golang.org/x/net/context"
)

// TopicGetRequest ...
type TopicGetRequest struct {
	TopicID int `json:"topicId"`
}

// TopicGetRequestDecode ...
func TopicGetRequestDecode(ctx context.Context, _ *http.Request) (interface{}, error) {
	topicID, err := rest.ParamFromContextInt(ctx, "topicId")
	if err != nil {
		return nil, err
	}

	return TopicGetRequest{
		TopicID: topicID,
	}, nil
}
