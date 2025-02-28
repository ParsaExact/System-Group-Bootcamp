package wordfrequency

import "strings"

func WordFrequency(text string) (res map[string]int) {

	words := strings.Fields(text)
	res = make(map[string]int)
	for _, word := range words {
		if word != "" {
			res[word]++
		}
	}
	return res
}
