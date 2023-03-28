package queryexpr

import (
	"context"
	"time"
)

type timeSource interface {
	now() time.Time
}

type defaultTimeSource struct{}

func (tds defaultTimeSource) now() time.Time {
	return time.Now()
}

type timeSourceContextKeyType struct{}

var timeSourceContextKey = timeSourceContextKeyType{}

func timeSourceFromContext(ctx context.Context) timeSource {
	if tts, ok := ctx.Value(timeSourceContextKey).(timeSource); ok {
		return tts
	}
	return defaultTimeSource{}
}
