package counter

import (
	"fmt"
)

func Counter(stepSize int) (func() int, error) {
	if stepSize <= 0 {
		return nil, fmt.Errorf("step should be positive; got: %d", stepSize)
	}

	count := -stepSize
	return func() int {
		count += stepSize
		return count
	}, nil
}
