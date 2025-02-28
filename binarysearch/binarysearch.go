package binarysearch

func BinarySearch(arr []string, key string) (index int) {
	low := 0
	high := len(arr) - 1

	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == key {
			return mid
		}
		if arr[mid] < key {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}
