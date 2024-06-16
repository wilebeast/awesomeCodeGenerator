package main

import "fmt"

func main() {
	result := calculate(1, 2)
	fmt.Println("Result:", result)
}

func calculate(a, b int) int {
	sum := add(a, b)
	diff := subtract(a, b)
	return sum * diff
}
