package ugly

func isPrime(n int) (ok bool) {
	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func IsUgly(n int) (ok bool) {
	if n <= 1 {
		return false
	}
    for i := 2; i <= n; i++ {
		if n%i == 0 && isPrime(i) && i != 2 && i != 3 && i != 5 {
			return false
		}
	}
	return true
}