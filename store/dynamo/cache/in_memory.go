package cache

import (
	"context"
	"time"

	"github.com/forta-network/core-go/store/dynamo"

	"github.com/patrickmn/go-cache"
)

type inMemory[I dynamo.Item] struct {
	cache *cache.Cache
}

// NewInMemoryWithTTL creates new in memory cache with TTL.
func NewInMemoryWithTTL[I dynamo.Item](expire time.Duration, checkEvery time.Duration) dynamo.Cache[I] {
	return &inMemory[I]{
		cache: cache.New(expire, checkEvery),
	}
}

func (cache *inMemory[I]) Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, bool) {
	v, ok := cache.cache.Get(makeCacheKey(partitionKey, sortKey...))
	if !ok {
		return nil, false
	}
	return v.(*I), true
}

func (cache *inMemory[I]) Put(ctx context.Context, item *I, partitionKey string, sortKey ...string) {
	cache.cache.Add(makeCacheKey(partitionKey, sortKey...), item, 0)
}

func (cache *inMemory[I]) GetQuery(ctx context.Context, queryKey string) ([]*I, bool) {
	v, ok := cache.cache.Get(queryKey)
	if !ok {
		return nil, false
	}
	return v.([]*I), true
}

func (cache *inMemory[I]) PutQuery(ctx context.Context, queryKey string, items []*I) {
	cache.cache.Add(queryKey, items, 0)
}
