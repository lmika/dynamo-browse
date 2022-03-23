package testdynamo

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecordAsItem(t *testing.T, item map[string]interface{}) models.Item {
	m, err := attributevalue.MarshalMap(item)
	assert.NoError(t, err)

	return models.Item(m)
}
