package maybe

type Maybe[T any] struct {
	val T
	err error
}

func Just[T any](val T) Maybe[T] {
	return Maybe[T]{val: val}
}

func Err[T any](err error) Maybe[T] {
	return Maybe[T]{err: err}
}

func (m Maybe[T]) Get() (T, error) {
	var zeroT T
	if m.err != nil {
		return zeroT, m.err
	}
	return m.val, nil
}
