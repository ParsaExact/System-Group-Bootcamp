package roman

func ToRoman(n int) string {
	a := n / 1000
	ans := ""
	for i := 0; i < a; i++ {
		ans += "M"
	}
	b := n / 100 % 10
	if b < 4 {
		for i := 0; i < b; i++ {
			ans += "C"
		}
	} else if b == 4 {
		ans += "CD"
	} else if b < 9 {
		ans += "D"
		for i := 0; i < b-5; i++ {
			ans += "C"
		}
	} else {
		ans += "CM"
	}
	c := n / 10 % 10
	if c < 4 {
		for i := 0; i < c; i++ {
			ans += "X"
		}
	} else if c == 4 {
		ans += "XL"
	} else if c < 9 {
		ans += "L"
		for i := 0; i < c-5; i++ {
			ans += "X"
		}
	} else {
		ans += "XC"
	}
	d := n % 10
	if d < 4 {
		for i := 0; i < d; i++ {
			ans += "I"
		}
	} else if d == 4 {
		ans += "IV"
	} else if d < 9 {
		ans += "V"
		for i := 0; i < d-5; i++ {
			ans += "I"
		}
	} else {
		ans += "IX"
	}
	return ans

}
