package main

import "fmt"

func main() {
    var n int
    fmt.Scanf("%d", &n)
	for i := 1; i <= n; i++ {
		for j := 1; j <= n - i; j++ {
			fmt.Print(" ")
		}
		for j := 1; j <= i; j++ {
			fmt.Print(j," ")
		}
		for k := i - 1; k > 0; k-- {
			fmt.Print(k," ")
		}
		fmt.Println()
	}
}