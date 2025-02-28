package dataset

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func PreProcessDataset(address string) (err error) {
	file, err := os.Open(address)
	if err != nil {
		return err
	}
	defer file.Close()

	var data [][]float64
	var columnsCount int
	lineNumber := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if lineNumber == 0 {
			columnsCount = len(parts)
		} else {
			if len(parts) != columnsCount {
				return fmt.Errorf("inconsistent column count in row %d", lineNumber+1)
			}
		}

		row := make([]float64, len(parts))
		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				return fmt.Errorf("empty value in row %d column %d", lineNumber+1, i+1)
			}
			val, parseErr := strconv.ParseFloat(trimmed, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid value %q in row %d column %d: %v", part, lineNumber+1, i+1, parseErr)
			}
			row[i] = val
		}

		data = append(data, row)
		lineNumber++
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	if len(data) == 0 {
		file, err := os.Create(address)
		if err != nil {
			return err
		}
		file.Close()
		return nil
	}

	numRows := len(data)
	numColumns := len(data[0])

	columns := make([][]float64, numColumns)
	for j := 0; j < numColumns; j++ {
		columns[j] = make([]float64, numRows)
		for i := 0; i < numRows; i++ {
			columns[j][i] = data[i][j]
		}
	}

	for j := range columns {
		col := columns[j]
		if len(col) == 0 {
			continue
		}

		min := col[0]
		max := col[0]
		for _, val := range col {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}

		for i, val := range col {
			if max == min {
				columns[j][i] = math.NaN()
			} else {
				columns[j][i] = (val - min) / (max - min)
			}
		}
	}

	processedRows := make([][]float64, numRows)
	for i := 0; i < numRows; i++ {
		processedRows[i] = make([]float64, numColumns)
		for j := 0; j < numColumns; j++ {
			processedRows[i][j] = columns[j][i]
		}
	}

	file, err = os.Create(address)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, row := range processedRows {
		var parts []string
		for _, val := range row {
			parts = append(parts, fmt.Sprintf("%.2f", val))
		}
		line := strings.Join(parts, ",")
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		return err
	}

	return nil
}
