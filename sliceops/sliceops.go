package sliceops

func RemoveAll(arr []string, key string) (res []string) {
	for _, elem := range arr {
		if elem != key {
			res = append(res, elem)
		}
	}
	return res
}

func DropAt(arr []string, index int) []string {
	for i := 0; i < len(arr); i++ {
		if i == index {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func InsertAt(arr []string, index int, elem string) []string {
	if index == 0 {
		return append([]string{elem}, arr...)
	}
	var left = arr[:index]
	var right = arr[index:]
	var cp_l = make([]string, len(left))
	copy(cp_l, left)
	cp_l = append(cp_l, elem)

	return append(cp_l, right...)
}
