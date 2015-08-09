package main

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// TopicsGetRequest ...
type TopicsGetRequest struct {
	Offset int
	Limit  int
}

// TopicsGetRequestDecode ...
func TopicsGetRequestDecode(ctx context.Context, r *http.Request) (interface{}, error) {
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	if err != nil {
		limit = 10
	}

	return TopicsGetRequest{
		Offset: int(offset),
		Limit:  int(limit),
	}, nil
}
