package main

import (
	"errors"

	"golang.org/x/net/context"
)

const (
	contextKeyTopicStorage = "topic_storage"
)

// NewTopicStorageContext returns a new Context that carries storage object.
func NewTopicStorageContext(ctx context.Context, storage *TopicStore) context.Context {
	return context.WithValue(ctx, contextKeyTopicStorage, storage)
}

// TopicStorageFromContext returns the storage stored in ctx, if any.
func TopicStorageFromContext(ctx context.Context) (*TopicStore, error) {
	s, ok := ctx.Value(contextKeyTopicStorage).(*TopicStore)

	if !ok {
		return nil, errors.New("missing storage in context")
	}

	return s, nil
}
