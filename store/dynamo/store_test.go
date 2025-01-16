package dynamo_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	mock_aws "github.com/forta-network/core-go/aws/mocks"
	"github.com/forta-network/core-go/store/dynamo"
	"github.com/forta-network/core-go/store/dynamo/cache"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const (
	testTableName       = "test-table"
	testIndexName       = "test-index"
	testPartitionKeyVal = "pkeyval"
	testSortKeyVal      = "skeyval"
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

var (
	testPartitionKeyOnly = map[string]types.AttributeValue{
		"pkey": &types.AttributeValueMemberS{Value: testPartitionKeyVal},
		"skey": &types.AttributeValueMemberS{Value: testSortKeyVal},
	}

	testBothKeys = map[string]types.AttributeValue{
		"pkey": &types.AttributeValueMemberS{Value: testPartitionKeyVal},
		"skey": &types.AttributeValueMemberS{Value: testSortKeyVal},
	}

	testFoundItem = testBothKeys

	testFoundItemBroken = map[string]types.AttributeValue{
		"pkey": &types.AttributeValueMemberNS{Value: []string{"1337"}},
	}
)

type primaryKeyMatcher struct {
	condition          *string
	expectedPrimaryKey map[string]types.AttributeValue
}

// Matches returns whether x is a match.
func (pkm *primaryKeyMatcher) Matches(x interface{}) bool {
	var key map[string]types.AttributeValue
	switch v := x.(type) {
	case *dynamodb.GetItemInput:
		key = v.Key

	case *dynamodb.PutItemInput:
		if pkm.condition != nil && *pkm.condition != *v.ConditionExpression {
			return false
		}
		key = v.Item

	case *dynamodb.DeleteItemInput:
		key = v.Key

	case *dynamodb.QueryInput:
		key = v.ExpressionAttributeValues

	default:
		panic(fmt.Sprintf("received unexpected type %T", x))
	}

	for keyName, keyValue := range key {
		expectedVal, ok := pkm.expectedPrimaryKey[keyName]
		if !ok {
			return false
		}
		s, ok := keyValue.(*types.AttributeValueMemberS)
		if !ok {
			return false
		}
		if s.Value != expectedVal.(*types.AttributeValueMemberS).Value {
			return false
		}
	}
	return true
}

// String describes what the matcher matches.
func (pkm *primaryKeyMatcher) String() string {
	b, _ := json.Marshal(pkm.expectedPrimaryKey)
	return string(b)
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name             string
		partitionKey     string
		sortKey          string
		addExpectedCalls func(store *mock_aws.MockDynamoDBClient)
		fail             bool
		expectErr        error
	}{
		{
			name:         "empty partition key",
			partitionKey: "",
			fail:         true,
		},
		{
			name:         "partition key only",
			partitionKey: testPartitionKeyVal,
			addExpectedCalls: func(store *mock_aws.MockDynamoDBClient) {
				store.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testPartitionKeyOnly}).
					Return(&dynamodb.GetItemOutput{
						Item: testFoundItem,
					}, nil)
			},
			fail: false,
		},
		{
			name:         "partition key and sort key",
			partitionKey: testPartitionKeyVal,
			sortKey:      testSortKeyVal,
			addExpectedCalls: func(store *mock_aws.MockDynamoDBClient) {
				store.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
					Return(&dynamodb.GetItemOutput{
						Item: testFoundItem,
					}, nil)
			},
			fail: false,
		},
		{
			name:         "not found",
			partitionKey: testPartitionKeyVal,
			sortKey:      testSortKeyVal,
			addExpectedCalls: func(store *mock_aws.MockDynamoDBClient) {
				store.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
					Return(&dynamodb.GetItemOutput{
						Item: nil,
					}, nil)
			},
			fail:      true,
			expectErr: dynamo.ErrNotFound,
		},
		{
			name:         "unknown error",
			partitionKey: testPartitionKeyVal,
			sortKey:      testSortKeyVal,
			addExpectedCalls: func(store *mock_aws.MockDynamoDBClient) {
				store.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
					Return(nil, errors.New("some unknown error"))
			},
			fail: true,
		},
		{
			name:         "unmarshal error",
			partitionKey: testPartitionKeyVal,
			sortKey:      testSortKeyVal,
			addExpectedCalls: func(store *mock_aws.MockDynamoDBClient) {
				store.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
					Return(&dynamodb.GetItemOutput{
						Item: testFoundItemBroken,
					}, nil)
			},
			fail: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
			if testCase.addExpectedCalls != nil {
				testCase.addExpectedCalls(client)
			}
			testItemStore := dynamo.NewStore[testItem](client, testTableName)
			var (
				item *testItem
				err  error
			)
			if len(testCase.sortKey) > 0 {
				item, err = testItemStore.Get(context.Background(), testCase.partitionKey, testCase.sortKey)
			} else {
				item, err = testItemStore.Get(context.Background(), testCase.partitionKey)
			}
			if testCase.fail {
				r.Error(err)
				if testCase.expectErr != nil {
					r.Equal(testCase.expectErr, err)
				}
				return
			}
			r.NoError(err)
			r.Equal(testPartitionKeyVal, item.Pkey)
			r.Equal(testSortKeyVal, item.Skey)
		})
	}
}

func TestGet_TempCache(t *testing.T) {
	r := require.New(t)

	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
	testItemStore := dynamo.NewStore[testItem](client, testTableName).WithCache(cache.NewTemp[testItem]())

	// expect only one call
	client.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
		Return(&dynamodb.GetItemOutput{
			Item: testFoundItem,
		}, nil)

	// first gets from the table, second from the cache
	expectedItem := &testItem{Pkey: testPartitionKeyVal, Skey: testSortKeyVal}
	retItem, err := testItemStore.Get(context.Background(), testPartitionKeyVal, testSortKeyVal)
	r.NoError(err)
	r.Equal(expectedItem, retItem)
	retItem, err = testItemStore.Get(context.Background(), testPartitionKeyVal, testSortKeyVal)
	r.NoError(err)
	r.Equal(expectedItem, retItem)
}

func TestGet_CacheLayer(t *testing.T) {
	r := require.New(t)

	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))

	// initialize the second store by layering on top of first but with a different cache
	testItemStore1 := dynamo.NewStore[testItem](client, testTableName).WithCache(cache.NewTemp[testItem]())
	testItemStore2 := testItemStore1.WithCache(cache.NewInMemoryWithTTL[testItem](time.Minute*60, time.Minute*60))

	// putting to both stores
	item := &testItem{Pkey: testPartitionKeyVal, Skey: testSortKeyVal}
	client.EXPECT().PutItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).Times(2)
	r.NoError(testItemStore1.Put(context.Background(), item))
	r.NoError(testItemStore2.Put(context.Background(), item))

	// expect only one get call
	client.EXPECT().GetItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).
		Return(&dynamodb.GetItemOutput{
			Item: testFoundItem,
		}, nil)

	// first gets from the table, second from the first store's cache
	expectedItem := &testItem{Pkey: testPartitionKeyVal, Skey: testSortKeyVal}
	retItem, err := testItemStore1.Get(context.Background(), testPartitionKeyVal, testSortKeyVal)
	r.NoError(err)
	r.Equal(expectedItem, retItem)
	retItem, err = testItemStore2.Get(context.Background(), testPartitionKeyVal, testSortKeyVal)
	r.NoError(err)
	r.Equal(expectedItem, retItem)
}

func TestPut(t *testing.T) {
	r := require.New(t)
	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
	testItemStore := dynamo.NewStore[testItem](client, testTableName)
	condition := dynamo.ConditionExpression{
		Expression: "test-condition",
	}
	client.EXPECT().PutItem(
		gomock.Any(),
		&primaryKeyMatcher{expectedPrimaryKey: testBothKeys, condition: &condition.Expression},
	).Return(nil, nil)
	r.NoError(testItemStore.Put(context.Background(), &testItem{
		Pkey: testPartitionKeyVal,
		Skey: testSortKeyVal,
	}, condition))
}

func TestDelete(t *testing.T) {
	r := require.New(t)
	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
	testItemStore := dynamo.NewStore[testItem](client, testTableName).WithCache(cache.NewTemp[testItem]())
	client.EXPECT().DeleteItem(gomock.Any(), &primaryKeyMatcher{expectedPrimaryKey: testBothKeys}).Return(nil, nil)
	r.NoError(testItemStore.Delete(context.Background(), &testItem{}, testPartitionKeyVal, testSortKeyVal))
}

func TestGetAll(t *testing.T) {
	r := require.New(t)
	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
	testItemStore := dynamo.NewStore[testItem](client, testTableName)
	client.EXPECT().Query(gomock.Any(), &primaryKeyMatcher{
		expectedPrimaryKey: map[string]types.AttributeValue{
			dynamo.AttributePartitionKey: &types.AttributeValueMemberS{Value: testPartitionKeyVal},
		},
	}).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			testFoundItem,
		},
	}, nil)
	retItems, err := testItemStore.GetAll(context.Background(), testPartitionKeyVal)
	r.NoError(err)
	r.Len(retItems, 1)
}

func TestGetAllFromIndex(t *testing.T) {
	r := require.New(t)
	client := mock_aws.NewMockDynamoDBClient(gomock.NewController(t))
	testItemStore := dynamo.NewStore[testItem](client, testTableName)
	client.EXPECT().Query(gomock.Any(), &primaryKeyMatcher{
		expectedPrimaryKey: map[string]types.AttributeValue{
			dynamo.AttributePartitionKey: &types.AttributeValueMemberS{Value: testPartitionKeyVal},
		},
	}).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			testFoundItem,
		},
	}, nil)
	retItems, err := testItemStore.GetAllFromIndex(context.Background(), testIndexName, "pkey", testPartitionKeyVal)
	r.NoError(err)
	r.Len(retItems, 1)
}
