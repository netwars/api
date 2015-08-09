package main

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/piotrkowalczuk/rest"
	"golang.org/x/net/context"
)

// TopicGetEndpoint ...
func TopicGetEndpoint(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(TopicGetRequest)
	if !ok {
		return nil, rest.InternalServerError(endpoint.ErrBadCast, internalServerErrorMessage, 0)
	}

	storage, err := TopicStorageFromContext(ctx)
	if err != nil {
		return nil, rest.InternalServerError(err, internalServerErrorMessage, 0)
	}

	return storage.GetOrRetrieve(req.TopicID)
}
