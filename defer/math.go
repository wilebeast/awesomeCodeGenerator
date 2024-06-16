package main

import "awesomeCodeGenerator/ellen"

func sub(a, b int) (c int) {
	defer func() {
		ellen.Printf("sub", map[string]interface{}{"a": a, "b": b}, map[string]interface{}{"c": c})
	}()
	subtract(a, b)
	c = subtract(a, b)
	return c
}

func add(a, b int) (X1 int) {
	defer func() {
		ellen.Printf("add", map[string]interface{}{"a": a, "b": b}, map[string]interface{}{"X1": X1})
	}()
	return a + b
}

func subtract(a, b int) (X1 int) {
	defer func() {
		ellen.Printf("subtract", map[string]interface{}{"a": a, "b": b}, map[string]interface{}{"X1": X1})
	}()
	return a - b
}

func nothing() {
	defer func() {
		ellen.Printf("nothing", map[string]interface{}{}, map[string]interface{}{})
	}()
}
