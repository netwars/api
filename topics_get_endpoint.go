package main

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/piotrkowalczuk/rest"
	"golang.org/x/net/context"
)

// TopicsGetEndpoint ...
func TopicsGetEndpoint(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(TopicsGetRequest)
	if !ok {
		return nil, rest.InternalServerError(endpoint.ErrBadCast, internalServerErrorMessage, 0)
	}

	storage, err := TopicStorageFromContext(ctx)
	if err != nil {
		return nil, rest.InternalServerError(err, internalServerErrorMessage, 0)
	}

	topics, err := storage.List(req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(topics))
	for _, topic := range topics {
		response = append(response, map[string]interface{}{
			"id":        topic.ID,
			"forumId":   topic.ForumID,
			"title":     topic.Title,
			"updatedAt": topic.UpdatedAt,
		})
	}
	return response, nil
}
