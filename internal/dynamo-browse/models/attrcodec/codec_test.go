package attrcodec_test

import (
	"bytes"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrcodec"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCodec(t *testing.T) {
	t.Run("should be able to encode and decode", func(t *testing.T) {
		scenarios := []struct {
			name string
			val  types.AttributeValue
		}{
			{name: "string", val: &types.AttributeValueMemberS{Value: "Hello world"}},
			{name: "empty string", val: &types.AttributeValueMemberS{Value: ""}},
			{name: "large string", val: &types.AttributeValueMemberS{Value: strings.Repeat("DynamoDB", 256)}},

			{name: "number", val: &types.AttributeValueMemberN{Value: "12345"}},
			{name: "large number", val: &types.AttributeValueMemberN{Value: "123456789012345678901234567890"}},

			{name: "true bool", val: &types.AttributeValueMemberBOOL{Value: true}},
			{name: "false bool", val: &types.AttributeValueMemberBOOL{Value: false}},

			{name: "true null", val: &types.AttributeValueMemberNULL{Value: true}},
			{name: "false null", val: &types.AttributeValueMemberNULL{Value: false}},

			{name: "bytes", val: &types.AttributeValueMemberB{Value: []byte{1, 2, 3, 4, 5}}},

			{name: "simple list", val: &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: "apple"},
				&types.AttributeValueMemberS{Value: "banana"},
				&types.AttributeValueMemberS{Value: "cherry"},
			}}},
			{name: "nested lists", val: &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "red apple"},
					&types.AttributeValueMemberS{Value: "green apple"},
				}},
				&types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "banana"},
					&types.AttributeValueMemberS{Value: "banana bread"},
					&types.AttributeValueMemberS{Value: "banana cake"},
				}},
				&types.AttributeValueMemberS{Value: "cherry"},
				&types.AttributeValueMemberS{Value: "can't make anything with cherries"},
			}}},

			{name: "simple map", val: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"alpha":   &types.AttributeValueMemberS{Value: "I am an apple"},
				"bravo":   &types.AttributeValueMemberN{Value: "123.45"},
				"charlie": &types.AttributeValueMemberS{Value: "things go here"},
			}}},
			{name: "nested maps", val: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"alpha": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "red apple"},
					&types.AttributeValueMemberS{Value: "green apple"},
				}},
				"bravo": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"good": &types.AttributeValueMemberS{Value: "stuff"},
					"is":   &types.AttributeValueMemberS{Value: "written"},
					"in":   &types.AttributeValueMemberS{Value: "the unit tests"},
				}},
				"coords": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"lat":  &types.AttributeValueMemberN{Value: "12.34"},
						"long": &types.AttributeValueMemberN{Value: "45.78"},
					}},
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"lat":  &types.AttributeValueMemberN{Value: "11.22"},
						"long": &types.AttributeValueMemberN{Value: "33.44"},
					}},
				}},
			}}},

			{name: "binary set", val: &types.AttributeValueMemberBS{Value: [][]byte{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			}}},
			{name: "number set", val: &types.AttributeValueMemberNS{Value: []string{
				"123",
				"456",
				"789",
			}}},
			{name: "string set", val: &types.AttributeValueMemberSS{Value: []string{
				"more",
				"string",
				"stuff",
			}}},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				bfr := new(bytes.Buffer)

				err := attrcodec.NewEncoder(bfr).Encode(scenario.val)
				assert.NoError(t, err)

				t.Logf("length = %v", bfr.Len())

				otherVal, err := attrcodec.NewDecoder(bfr).Decode()
				assert.NoError(t, err)
				assert.Equal(t, scenario.val, otherVal)
			})
		}
	})
}
