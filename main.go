package main

import (
	"fmt"

	Array "github.com/go-composites/array/src"
	Enumerator "github.com/go-composites/enumerator/src"
	Result "github.com/go-composites/result/src"
)

func main() {
	// A finite prefix of an INFINITE sequence — lazily.
	evens := Enumerator.
		Generate(0, func(x interface{}) interface{} { return x.(int) + 1 }).
		Filter(func(x interface{}) bool { return x.(int)%2 == 0 }).
		Map(func(x interface{}) interface{} { return x.(int) * x.(int) }).
		Take(5).
		ToArray()
	fmt.Println("first 5 even squares:", show(evens)) // [0 4 16 36 64]

	sum := Enumerator.New(1, 2, 3, 4).Reduce(0, func(acc, item interface{}) Result.Interface {
		return Result.New(Result.WithPayload(acc.(int) + item.(int)))
	})
	fmt.Println("sum 1..4:", sum.Payload()) // 10

	fmt.Println("first of empty has error:", Enumerator.New().First().HasError()) // true
}

func show(arr Array.Interface) []interface{} {
	out := make([]interface{}, 0, arr.Len())
	for i := 0; i < arr.Len(); i++ {
		out = append(out, arr.Fetch(i).Payload())
	}
	return out
}
