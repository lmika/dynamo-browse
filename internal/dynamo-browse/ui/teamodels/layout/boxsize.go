package layout

type BoxSize interface {
	childSize(idx, cnt, available int) int
}

func EqualSize() BoxSize {
	return equalSize{}
}

type equalSize struct {
}

func (l equalSize) childSize(idx, cnt, available int) int {
	childrenHeight := available / cnt
	lastChildRem := available % cnt
	if idx == cnt-1 {
		return childrenHeight + lastChildRem
	}
	return childrenHeight
}

func LastChildFixedAt(size int) BoxSize {
	return lastChildFixedAt{size}
}

type lastChildFixedAt struct {
	lastChildSize int
}

func (l lastChildFixedAt) childSize(idx, cnt, available int) int {
	if idx == cnt-1 {
		return l.lastChildSize
	}
	return (equalSize{}).childSize(idx, cnt-1, available-l.lastChildSize)
}
