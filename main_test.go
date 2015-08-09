package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"time"

	"strconv"

	"github.com/netwars/api/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestTopicGetHandler(t *testing.T) {
	now := time.Now()
	success := map[int]*Topic{
		1: {
			ID:        1,
			ForumID:   1,
			Title:     "test",
			UpdatedAt: &now,
		},
		2: {
			ID:        111241241,
			ForumID:   235352,
			Title:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			UpdatedAt: &now,
		},
	}

	// ------------
	// ---Server---
	// ------------
	client := &ClientMock{}
	for id, topic := range success {
		client.On("FetchTopic", id).Return(topic, nil)
	}
	server := setupTestServer(client)
	defer server.Close()

	// ------------
	// ----Test----
	// ------------
	for id, topic := range success {
		res, err := http.Get(server.URL + "/topic/" + strconv.FormatInt(int64(id), 10))
		if !assert.NoError(t, err) {
			return
		}

		var requestedTopic *Topic

		err = json.NewDecoder(res.Body).Decode(&requestedTopic)
		if assert.NoError(t, err) {
			assert.Equal(t, requestedTopic.ID, topic.ID)
			assert.Equal(t, requestedTopic.ForumID, topic.ForumID)
			assert.Equal(t, requestedTopic.Title, topic.Title)
			assert.Equal(t, requestedTopic.Posts, topic.Posts)
			assert.Equal(t, requestedTopic.UpdatedAt, topic.UpdatedAt)
		}
	}
}

func setupTestServer(client Client) *httptest.Server {
	topicCache := cache.NewCache(cache.CacheOpts{
		Expiration: 1000000 * time.Hour,
		Interval:   1000000 * time.Hour,
	})
	topicStorage := NewTopicStore(client, topicCache, TopicStoreOpts{
		WarmUp: warmUp,
	})
	go logErrorChannel("topic-storage", topicStorage.Err())

	ctx := context.Background()
	ctx = NewTopicStorageContext(ctx, topicStorage)

	return httptest.NewServer(buildRoutes(ctx))
}

type ClientMock struct {
	mock.Mock
}

func (cm *ClientMock) FetchTopic(id int) (*Topic, error) {
	args := cm.Called(id)
	return args.Get(0).(*Topic), args.Error(1)
}

func (cm *ClientMock) FetchTopics(int, chan<- *Topic) error {
	return nil
}
