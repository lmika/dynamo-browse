package queryexpr

func (a *astExpr) String() string {
	return a.Equality.String()
}

func (a *astBinOp) String() string {
	return a.Name + a.Op + a.Value.String()
}

func (a *astLiteralValue) String() string {
	return a.StringVal
}
