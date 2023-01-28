package sliceutils

func All[T any](ts []T, predicate func(t T) bool) bool {
	for _, t := range ts {
		if !predicate(t) {
			return false
		}
	}
	return true
}

func Map[T, U any](ts []T, fn func(t T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = fn(t)
	}
	return us
}

func MapWithError[T, U any](ts []T, fn func(t T) (U, error)) ([]U, error) {
	us := make([]U, len(ts))

	for i, t := range ts {
		var err error
		us[i], err = fn(t)
		if err != nil {
			return nil, err
		}
	}
	return us, nil
}

func Filter[T any](ts []T, fn func(t T) bool) []T {
	us := make([]T, 0)
	for _, t := range ts {
		if fn(t) {
			us = append(us, t)
		}
	}
	return us
}
