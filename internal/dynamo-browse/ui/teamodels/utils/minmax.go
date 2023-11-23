package utils

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Cycle(n int, by int, max int) int {
	by = by % max
	if by > 0 {
		return (n + by) % max
	} else if by < 0 {
		wn := n + by
		if wn < 0 {
			return max + wn
		}
		return wn
	}
	return n
}
