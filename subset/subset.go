package subset

func IsSubsetOf(s1 []int, s2 []int) bool {
	if len(s1) > len(s2) {
		return false
	}
	items := make(map[int]bool)
	for _, v2 := range s2 {
		items[v2] = true
	}

	for _, v1 := range s1 {
		if _, ok := items[v1]; !ok {
			return false
		}
	}
	return true

}
