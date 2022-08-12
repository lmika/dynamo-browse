package queryexpr

import "strings"

func (a *astExpr) String() string {
	return a.Root.String()
}

func (a *astDisjunction) String() string {
	sb := new(strings.Builder)
	for i, op := range a.Operands {
		if i > 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString(op.String())
	}
	return sb.String()
}

func (a *astBinOp) String() string {
	return a.Name + a.Op + a.Value.String()
}

func (a *astLiteralValue) String() string {
	return a.StringVal
}
