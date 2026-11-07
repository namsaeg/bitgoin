package transaction

// cloneAll은 슬라이스 내의 모든 요소를 Clone() 메서드로 복사한 새 슬라이스를 반환합니다.
func cloneAll[T interface{ Clone() T }](src []T) []T {
	dst := make([]T, len(src))
	for i, v := range src {
		dst[i] = v.Clone()
	}
	return dst
}
