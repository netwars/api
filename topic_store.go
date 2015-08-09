package main

import (
	"errors"
	"log"
	"sort"

	"github.com/netwars/api/cache"
)

// TopicStoreOpts ...
type TopicStoreOpts struct {
	WarmUp int
}

// TopicStore ...
type TopicStore struct {
	cache.Cache
	err          chan error
	client       Client
	index        []int
	notification chan int
}

// NewTopicStore ...
func NewTopicStore(client Client, cache *cache.Cache, options TopicStoreOpts) *TopicStore {
	store := &TopicStore{
		Cache:  *cache,
		client: client,
		err:    make(chan error, 1),
		index:  make([]int, 0, 10000), // made up value
	}

	go store.listenCache()

	if options.WarmUp > 0 {
		go store.warmUp(options.WarmUp)
	}

	return store
}

// Set ...
func (ts *TopicStore) Set(topic *Topic) {
	ts.Cache.Set(topic.ID, topic)

	ts.Lock()
	defer ts.Unlock()
	if !ts.indexed(topic.ID) {
		ts.index = append(ts.index, topic.ID)
	}

	// re-index anyway, value can be different
	ts.ReIndex()

	return
}

func (ts *TopicStore) indexed(id int) bool {
	for _, i := range ts.index {
		if i == id {
			return true
		}
	}

	return false
}

func (ts *TopicStore) listenCache() {
	for {
		select {
		case id, open := <-ts.Cache.Notify():
			if !open {
				return
			}

			topic, err := ts.client.FetchTopic(id)
			if err != nil {
				ts.err <- err
				continue
			}

			ts.Set(topic)
			log.Printf("[%d] update notification - topic fetched and updated successfully: %s", topic.ID, topic.Title)
		case e, open := <-ts.Cache.Err():
			if !open {
				return
			}
			ts.err <- e
		}
	}
}

// ReIndex is not thread safe!
func (ts *TopicStore) ReIndex() {
	sort.Sort(ts)
}

// Len ...
func (ts *TopicStore) Len() int {
	return len(ts.index)
}

// Swap is not thread safe!
func (ts *TopicStore) Swap(i, j int) {
	ts.index[i], ts.index[j] = ts.index[j], ts.index[i]
}

// Less is not thread safe!
func (ts *TopicStore) Less(i, j int) bool {
	ti, ok1 := ts.Get(ts.index[i]).(*Topic)
	tj, ok2 := ts.Get(ts.index[j]).(*Topic)

	if !ok1 || !ok2 {
		ts.err <- errors.New("cache contains object that cannot be casted to Topic")
		return false
	}

	return ti.UpdatedAt.Before(*tj.UpdatedAt)
}

// Err ...
func (ts *TopicStore) Err() <-chan error {
	return ts.err
}

// List ...
func (ts *TopicStore) List(offset, limit int) ([]*Topic, error) {
	if limit == 0 {
		return []*Topic{}, nil
	}

	ts.RLock()
	defer ts.RUnlock()

	if offset > len(ts.index) {
		return nil, errors.New("offset out of range")
	}

	if limit > len(ts.index) {
		limit = len(ts.index)
	}

	topics := make([]*Topic, 0, limit-offset)

	for i := limit - 1; i >= offset; i-- {
		topics = append(topics, ts.Get(ts.index[i]).(*Topic))
	}

	return topics, nil
}

func (ts *TopicStore) warmUp(warmUp int) {
	topics := make(chan *Topic)

	go func() {
		for topic := range topics {
			ts.Set(topic)
			log.Printf("[%d] warmup - topic fetched and updated successfully: %s", topic.ID, topic.Title)
		}
	}()

	if err := ts.client.FetchTopics(warmUp, topics); err != nil {
		close(topics)
		ts.err <- err

		return
	}
}

// GetOrRetrieve ...
func (ts *TopicStore) GetOrRetrieve(id int) (*Topic, error) {
	var err error

	topic, ok := ts.SafeGet(id).(*Topic)
	if !ok {
		topic, err = ts.client.FetchTopic(id)
		if err != nil {
			return nil, err
		}

		ts.Set(topic)
	}

	return topic, nil
}
