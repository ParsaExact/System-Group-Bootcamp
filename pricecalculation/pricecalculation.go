package pricecalculation

func CalculatePrice(prices map[string]int, cart []string) (price int) {
	var ans = 0

	for _, item := range cart {
		for key, _ := range prices {
			if price, ok := prices[key]; ok && key == item {
				ans += price
			}
		}

	}
	return ans
}
