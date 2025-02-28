package cachedfib

var fibCache = make(map[int]int64)

func CachedFib() func(int) int64 {
	return func(n int) int64 {
		if val, ok := fibCache[n]; ok {
			return val
		}

		if n <= 1 {
			return int64(n)
		}

		result := CachedFib()(n-1) + CachedFib()(n-2)
		fibCache[n] = result
		return result
	}
}
