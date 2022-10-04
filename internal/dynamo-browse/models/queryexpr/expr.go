package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

type QueryExpr struct {
	ast *astExpr
}

func (md *QueryExpr) Plan(tableInfo *models.TableInfo) (*models.QueryExecutionPlan, error) {
	ir, err := md.ast.evalToIR(tableInfo)
	if err != nil {
		return nil, err
	}

	return ir.calcQuery(tableInfo)
}

func (md *QueryExpr) EvalItem(item models.Item) (types.AttributeValue, error) {
	return md.ast.evalItem(item)
}

func (md *QueryExpr) String() string {
	return md.ast.String()
}

func (a *astExpr) String() string {
	return a.Root.String()
}

type queryCalcInfo struct {
	seenKeys map[string]struct{}
}

func (qc *queryCalcInfo) addKey(tableInfo *models.TableInfo, key string) bool {
	if tableInfo.Keys.PartitionKey != key && tableInfo.Keys.SortKey != key {
		return false
	}

	if qc.seenKeys == nil {
		qc.seenKeys = make(map[string]struct{})
	}
	if _, hasSeenKey := qc.seenKeys[key]; hasSeenKey {
		return false
	}

	qc.seenKeys[key] = struct{}{}
	return true
}
