package cache

import (
	"context"

	"github.com/forta-network/core-go/store/dynamo"
)

type tempCache[I dynamo.Item] map[string]interface{}

// NewTemp creates a new temp cache.
func NewTemp[I dynamo.Item]() dynamo.Cache[I] {
	return tempCache[I]{}
}

func (cache tempCache[I]) Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, bool) {
	it, ok := cache[makeCacheKey(partitionKey, sortKey...)]
	if !ok {
		return nil, false
	}
	return it.(*I), true
}

func (cache tempCache[I]) Put(ctx context.Context, item *I, partitionKey string, sortKey ...string) {
	cache[makeCacheKey(partitionKey, sortKey...)] = item
}

func (cache tempCache[I]) GetQuery(ctx context.Context, queryKey string) ([]*I, bool) {
	its, ok := cache[queryKey]
	if !ok {
		return nil, false
	}
	return its.([]*I), ok
}

func (cache tempCache[I]) PutQuery(ctx context.Context, queryKey string, items []*I) {
	cache[queryKey] = items
}
