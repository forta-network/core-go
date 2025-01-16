package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testItem struct {
	Pkey string `dynamodbav:"pkey"`
	Skey string `dynamodbav:"skey"`
}

func (item testItem) GetPartitionKeyName() string {
	return "pkey"
}

func (item testItem) GetSortKeyName() string {
	return "skey"
}

func TestInMemory(t *testing.T) {
	r := require.New(t)

	cache := NewInMemoryWithTTL[testItem](time.Minute*60, time.Minute*60)

	it, ok := cache.Get(context.Background(), "pkeyval", "skeyval")
	r.Nil(it)
	r.False(ok)

	cache.Put(context.Background(), &testItem{}, "pkeyval", "skeyval")

	it, ok = cache.Get(context.Background(), "pkeyval", "skeyval")
	r.NotNil(it)
	r.True(ok)

	its, ok := cache.GetQuery(context.Background(), "querykey")
	r.Nil(its)
	r.False(ok)

	cache.PutQuery(context.Background(), "querykey", []*testItem{{}})

	its, ok = cache.GetQuery(context.Background(), "querykey")
	r.NotNil(its)
	r.True(ok)
}
