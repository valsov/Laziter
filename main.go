package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	iter := New[int](true)
	vp := iter.GetValuesProvider()
	go test(vp)

	for iter.Next() {
		val, _ := iter.GetCurrentValue()
		fmt.Println(val)
	}
	fmt.Println(time.Since(start))

	start = time.Now()
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
	fmt.Println(time.Since(start))
}

func test(vp ValuesProvider[int]) {
	defer vp.Close()
	for i := 0; i < 10; i++ {
		if !vp.Yield(i) {
			return
		}
	}
}
