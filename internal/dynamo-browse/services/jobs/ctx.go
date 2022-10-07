package jobs

import (
	"context"
)

type jobUpdaterKeyType struct{}

var jobUpdaterKey = jobUpdaterKeyType{}

type jobUpdaterValue struct {
	msgUpdate chan string
}

func PostUpdate(ctx context.Context, msg string) {
	val, hasVal := ctx.Value(jobUpdaterKey).(*jobUpdaterValue)
	if !hasVal {
		return
	}

	select {
	case val.msgUpdate <- msg:
	default:
	}
}
