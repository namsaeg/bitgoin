package clone

func CloneAll[T interface{ Clone() T }](src []T) []T {
	dst := make([]T, len(src))
	for i, v := range src {
		dst[i] = v.Clone()
	}
	return dst
}
