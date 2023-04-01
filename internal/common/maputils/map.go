package maputils

func Values[K comparable, T any](ts map[K]T) []T {
	values := make([]T, 0, len(ts))
	for _, v := range ts {
		values = append(values, v)
	}
	return values
}

func MapValues[K comparable, T, U any](ts map[K]T, fn func(t T) U) map[K]U {
	us := make(map[K]U)

	for k, t := range ts {
		us[k] = fn(t)
	}

	return us
}

func MapValuesWithError[K comparable, T, U any](ts map[K]T, fn func(t T) (U, error)) (map[K]U, error) {
	us := make(map[K]U)

	for k, t := range ts {
		var err error
		us[k], err = fn(t)
		if err != nil {
			return nil, err
		}
	}

	return us, nil
}
