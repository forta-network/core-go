package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/forta-network/core-go/aws"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Attribute keys
const (
	AttributePartitionKey = ":partitionkeyval"
)

// Item represents the minimum interface that needs to be implemented
// by the items that are managed by the store.
type Item interface {
	GetPartitionKeyName() string
	GetSortKeyName() string
}

type ConditionExpression struct {
	Expression                string
	ExpressionAttributeValues map[string]types.AttributeValue
}

// Cache is a generic interface for caching data.
type Cache[I Item] interface {
	Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, bool)
	Put(ctx context.Context, item *I, partitionKey string, sortKey ...string)
	GetQuery(ctx context.Context, queryKey string) ([]*I, bool)
	PutQuery(ctx context.Context, queryKey string, items []*I)
}

// Store is a generic interface for storing data.
type Store[I Item] interface {
	TableName() string
	WithCache(cache Cache[I]) Store[I]

	Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, error)
	GetAll(ctx context.Context, partitionKey string) ([]*I, error)
	GetAllFromIndex(ctx context.Context, indexName, partitionKeyName, partitionKeyVal string) ([]*I, error)
	Put(ctx context.Context, item *I, conditionExpression ...ConditionExpression) error
	Delete(ctx context.Context, item *I, partitionKey string, sortKey ...string) error
}

// Store errors
var (
	ErrNotFound = errors.New("not found")
)

type store[I Item] struct {
	client    aws.DynamoDBClient
	tableName string
	item      I
}

// NewStore creates a new store.
func NewStore[I Item](client aws.DynamoDBClient, tableName string) Store[I] {
	return &store[I]{
		client:    client,
		tableName: tableName,
	}
}

func (s *store[I]) TableName() string {
	return s.tableName
}

func (s *store[I]) WithCache(cache Cache[I]) Store[I] {
	return &cachedStore[I]{cache: cache, Store: s}
}

func (s *store[I]) Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, error) {
	if len(partitionKey) == 0 {
		return nil, errors.New("empty partition key provided")
	}

	var item I
	res, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key:       makePrimaryKey(&item, partitionKey, sortKey...),
	})
	if res != nil && res.Item == nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := attributevalue.UnmarshalMap(res.Item, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func makePrimaryKey[I Item](item *I, partitionKey string, sortKey ...string) map[string]types.AttributeValue {
	primaryKey := make(map[string]types.AttributeValue, 2)

	primaryKey[(*item).GetPartitionKeyName()] = &types.AttributeValueMemberS{Value: partitionKey}
	// add the sort key only if non-empty value is provided
	if sortKey != nil {
		primaryKey[(*item).GetSortKeyName()] = &types.AttributeValueMemberS{Value: sortKey[0]}
	}
	return primaryKey
}

func (s *store[I]) Put(ctx context.Context, item *I, conditionExpression ...ConditionExpression) error {
	marshaled, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	op := &dynamodb.PutItemInput{
		Item:      marshaled,
		TableName: &s.tableName,
	}
	if conditionExpression != nil {
		op.ConditionExpression = &conditionExpression[0].Expression
		op.ExpressionAttributeValues = conditionExpression[0].ExpressionAttributeValues
	}
	_, err = s.client.PutItem(ctx, op)
	return err
}

func (s *store[I]) Delete(ctx context.Context, item *I, partitionKey string, sortKey ...string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key:       makePrimaryKey(item, partitionKey, sortKey...),
	})
	return err
}

func (s *store[I]) GetAll(ctx context.Context, partitionKeyVal string) ([]*I, error) {
	keyCond := fmt.Sprintf("%s = %s", s.item.GetPartitionKeyName(), AttributePartitionKey)
	res, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: &keyCond,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			AttributePartitionKey: &types.AttributeValueMemberS{Value: partitionKeyVal},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all with same partition key: %v", err)
	}
	var items []*I
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *store[I]) GetAllFromIndex(ctx context.Context, indexName, partitionKeyName, partitionKeyVal string) ([]*I, error) {
	keyCond := fmt.Sprintf("%s = %s", partitionKeyName, AttributePartitionKey)
	res, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &s.tableName,
		IndexName:              &indexName,
		KeyConditionExpression: &keyCond,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			AttributePartitionKey: &types.AttributeValueMemberS{Value: partitionKeyVal},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all with same partition key: %v", err)
	}
	var items []*I
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &items); err != nil {
		return nil, err
	}
	return items, nil
}

type cachedStore[I Item] struct {
	cache Cache[I]
	Store[I]
}

func (s *cachedStore[I]) Get(ctx context.Context, partitionKey string, sortKey ...string) (*I, error) {
	if s.cache != nil {
		it, ok := s.cache.Get(ctx, partitionKey, sortKey...)
		if ok {
			return it, nil
		}
	}

	it, err := s.Store.Get(ctx, partitionKey, sortKey...)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.Put(ctx, it, partitionKey, sortKey...)
	}

	return it, nil
}

func (s *cachedStore[I]) Put(ctx context.Context, item *I, conditionExpression ...ConditionExpression) error {
	return s.Store.Put(ctx, item, conditionExpression...)
}

func (s *cachedStore[I]) GetAll(ctx context.Context, partitionKey string) ([]*I, error) {
	cacheKey := makeQueryCacheKey(s.TableName(), partitionKey)

	if s.cache != nil {
		its, ok := s.cache.GetQuery(ctx, cacheKey)
		if ok {
			return its, nil
		}
	}

	its, err := s.Store.GetAll(ctx, partitionKey)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.PutQuery(ctx, cacheKey, its)
	}

	return its, nil
}

func (s *cachedStore[I]) GetAllFromIndex(ctx context.Context, indexName, partitionKeyName, partitionKeyVal string) ([]*I, error) {
	cacheKey := makeQueryCacheKey(indexName, partitionKeyName, partitionKeyVal)

	if s.cache != nil {
		its, ok := s.cache.GetQuery(ctx, cacheKey)
		if ok {
			return its, nil
		}
	}

	its, err := s.Store.GetAllFromIndex(ctx, indexName, partitionKeyName, partitionKeyVal)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.PutQuery(ctx, cacheKey, its)
	}

	return its, nil
}

func makeQueryCacheKey(values ...string) string {
	return strings.Join(values, "@@")
}

func (s *cachedStore[I]) Delete(ctx context.Context, item *I, partitionKey string, sortKey ...string) error {
	return s.Store.Delete(ctx, item, partitionKey, sortKey...)
}

func (s *cachedStore[I]) WithCache(cache Cache[I]) Store[I] {
	return &cachedStore[I]{
		cache: cache,
		Store: s, // this allows layering cached stores
	}
}
