package inversion

import "sort"

func InvertMap(input map[string]string) (output map[string][]string) {
	output = make(map[string][]string)
	for key, value := range input {
		if _, ok := output[value]; !ok {
			output[value] = []string{key}
		} else {
			output[value] = append(output[value], key)
			sort.Strings(output[value])
		}
	}
	return output
}
