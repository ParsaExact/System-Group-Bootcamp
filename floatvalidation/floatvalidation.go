package floatvalidation
import "math"
func ValidateFloat(value float64, min float64, max float64, precision int) (ok bool) {
	if min > max {	
		return false
	}
    if value < min || value > max {
		return false
	}
	tavan := math.Pow(10, float64(precision))
	s := tavan * value
	if s != float64(int(s)) {
		return false
	}
	return true
}