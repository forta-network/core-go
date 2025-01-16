package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemp(t *testing.T) {
	r := require.New(t)

	cache := NewTemp[testItem]()

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
