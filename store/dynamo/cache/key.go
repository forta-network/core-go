package cache

func makeCacheKey(partitionKey string, sortKey ...string) string {
	cacheKey := partitionKey
	if sortKey != nil {
		cacheKey += "|" + sortKey[0]
	}
	return cacheKey
}
