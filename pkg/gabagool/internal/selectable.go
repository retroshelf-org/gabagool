package internal

// NextSelectable is the nearest index from `from`, stepping by step (+1 or
// -1), that selectable accepts, wrapping past either end when wrap is set.
// `from` itself counts. It returns -1 when no index is selectable.
func NextSelectable(n int, selectable func(int) bool, from, step int, wrap bool) int {
	if n == 0 || step == 0 {
		return -1
	}
	i := from
	for range n {
		if wrap {
			i = ((i % n) + n) % n
		} else if i < 0 || i >= n {
			return -1
		}
		if selectable(i) {
			return i
		}
		i += step
	}
	return -1
}
