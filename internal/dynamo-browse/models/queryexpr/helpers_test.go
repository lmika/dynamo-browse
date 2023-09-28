package queryexpr

import (
	"time"
)

type testTimeSource time.Time

func (tds testTimeSource) now() time.Time {
	return time.Time(tds)
}

func (a *QueryExpr) WithTestTimeSource(timeNow time.Time) *QueryExpr {
	a.timeSource = testTimeSource(timeNow)
	return a
}

