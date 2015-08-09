package cache_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/netwars/api/cache"
	"github.com/stretchr/testify/assert"
)

const (
	benchmarkValue = "value"
)

var (
	benchmarkResult interface{}
)

func TestCache(t *testing.T) {
	n := int(500)
	data := make([]interface{}, 0, n)
	for i := int(0); i < n; i++ {
		data = append(data, fmt.Sprintf("forum%d", i))
	}

	options := cache.CacheOpts{
		Expiration: 100000 * time.Second,
		Interval:   1 * time.Second,
	}
	ca := cache.NewCache(options)

	go func() {
		for {
			select {
			case <-ca.Notify():
			case err, closed := <-ca.Err():
				if !closed {
					assert.Nil(t, err)
					continue
				}

				if assert.NotNil(t, err) {
					assert.Fail(t, "cache error should be never received, got: %v", err)
				} else {
					t.Log("cache didnt returned any errors")
				}
			}
		}
	}()

	for key, value := range data {
		ca.Set(int(key), value)
	}

	wg := sync.WaitGroup{}
	for i, expected := range data {
		wg.Add(1)
		go func(i int, expected interface{}) {
			defer wg.Done()
			assert.Equal(t, expected, ca.Get(i))
		}(i, expected)
	}
	wg.Wait()

	ca.Terminate()

	for i, expected := range data {
		wg.Add(1)
		go func(i int, expected interface{}) {
			defer wg.Done()
			assert.Nil(t, ca.Get(i))
		}(i, expected)
	}
	wg.Wait()
}

func BenchmarkCacheSet(b *testing.B) {
	b.Log("bench")
	options := cache.CacheOpts{
		Expiration: 100000 * time.Second,
		Interval:   100000 * time.Second,
	}
	ca := cache.NewCache(options)

	go func() {
		for {
			select {
			case <-ca.Notify():
			case <-ca.Err():

			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ca.Set(i, benchmarkValue)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	var r interface{}

	options := cache.CacheOpts{
		Expiration: 100000 * time.Second,
		Interval:   100000 * time.Second,
	}
	ca := cache.NewCache(options)

	go func() {
		for {
			select {
			case <-ca.Notify():
			case <-ca.Err():

			}
		}
	}()

	for i := 0; i < b.N; i++ {
		ca.Set(i, benchmarkValue)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = ca.Get(i)
	}

	benchmarkResult = r
}
