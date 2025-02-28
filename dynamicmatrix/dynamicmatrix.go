package dynamicmatrix

func SetElement(matrix [][]int, i int, j int, value int) (result [][]int) {
	row := max(len(matrix), i+1)
	col := 0
	if len(matrix) > 0 {
		col = max(len(matrix[0]), j+1)
	} else {
		col = j + 1
	}

	result = make([][]int, row)
	for k := 0; k < row; k++ {
		result[k] = make([]int, col)
	}

	for k := 0; k < row; k++ {
		for l := 0; l < col; l++ {
			if k == i && l == j {
				result[k][l] = value
			} else if k < len(matrix) && l < len(matrix[k]) {
				result[k][l] = matrix[k][l]
			}
		}
	}

	return result
}
