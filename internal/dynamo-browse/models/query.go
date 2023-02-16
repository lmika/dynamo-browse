package models

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/lmika/audax/internal/dynamo-browse/models/itemrender"
)

type QueryExecutionPlan struct {
	CanQuery   bool
	IndexName  string
	Expression expression.Expression
}

func (qep QueryExecutionPlan) Describe(dp DescribingPrinter) {
	if qep.CanQuery {
		dp.Println("  execute as: query")
	} else {
		dp.Println("  execute as: scan")
	}

	if qep.IndexName != "" {
		dp.Printf("  index: %v", qep.IndexName)
	}
	if keyCond := aws.ToString(qep.Expression.KeyCondition()); keyCond != "" {
		dp.Printf("  key condition: %v", keyCond)
	}
	if cond := aws.ToString(qep.Expression.Condition()); cond != "" {
		dp.Printf("  condition: %v", cond)
	}
	if filter := aws.ToString(qep.Expression.Filter()); filter != "" {
		dp.Printf("  filter: %v", filter)
	}
	if names := qep.Expression.Names(); len(names) > 0 {
		dp.Println("  names:")
		for k, v := range names {
			dp.Printf("    %v = %v", k, v)
		}
	}
	if values := qep.Expression.Values(); len(values) > 0 {
		dp.Println("  values:")
		for k, v := range values {
			r := itemrender.ToRenderer(v)
			dp.Printf("    %v (%v) = %v", k, r.TypeName(), r.StringValue())
		}
	}
}

type DescribingPrinter interface {
	Println(v ...any)
	Printf(format string, v ...any)
}
