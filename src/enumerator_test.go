package Enumerator_test

import (
	"testing"

	Array "github.com/go-composites/array/src"
	Enumerator "github.com/go-composites/enumerator/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

func toSlice(arr Array.Interface) []interface{} {
	out := make([]interface{}, 0, arr.Len())
	for i := 0; i < arr.Len(); i++ {
		out = append(out, arr.Fetch(i).Payload())
	}
	return out
}

func equal(a []interface{}, b ...interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ok(payload interface{}) Result.Interface {
	return Result.New(Result.WithPayload(payload))
}

func boom() Result.Interface {
	return Result.New(Result.WithError(Error.New("boom")))
}

func TestNewToArray(t *testing.T) {
	if !equal(toSlice(Enumerator.New(1, 2, 3).ToArray()), 1, 2, 3) {
		t.Fatal("New/ToArray")
	}
}

// TestLazyGenerateTake is the laziness proof: an infinite Generate source must
// terminate once Take has pulled its prefix. If Take were eager this hangs.
func TestLazyGenerateTake(t *testing.T) {
	got := Enumerator.Generate(1, func(x interface{}) interface{} { return x.(int) + 1 }).
		Take(3).ToArray()
	if !equal(toSlice(got), 1, 2, 3) {
		t.Fatalf("Generate/Take = %v", toSlice(got))
	}
}

func TestMapFilter(t *testing.T) {
	got := Enumerator.New(1, 2, 3, 4).
		Filter(func(x interface{}) bool { return x.(int)%2 == 0 }).
		Map(func(x interface{}) interface{} { return x.(int) * 10 }).
		ToArray()
	if !equal(toSlice(got), 20, 40) {
		t.Fatalf("Filter/Map = %v", toSlice(got))
	}
}

func TestTakeBounds(t *testing.T) {
	if Enumerator.New(1, 2, 3).Take(0).ToArray().Len() != 0 {
		t.Fatal("Take(0)")
	}
	if Enumerator.New(1, 2, 3).Take(-1).ToArray().Len() != 0 {
		t.Fatal("Take(-1)")
	}
	// Take fewer than available: stops via count == n (does not yield 3).
	if !equal(toSlice(Enumerator.New(1, 2, 3).Take(2).ToArray()), 1, 2) {
		t.Fatal("Take(2)")
	}
	// Take more than available: upstream exhausts first.
	if !equal(toSlice(Enumerator.New(1, 2).Take(5).ToArray()), 1, 2) {
		t.Fatal("Take(5)")
	}
	// Downstream stop (First) propagates ok=false up through Take.
	if Enumerator.New(1, 2, 3).Take(5).First().Payload() != 1 {
		t.Fatal("Take then First")
	}
}

func TestEach(t *testing.T) {
	sum := 0
	r := Enumerator.New(1, 2, 3).Each(func(item interface{}) Result.Interface {
		sum += item.(int)
		return Result.New()
	})
	if r.HasError() || sum != 6 {
		t.Fatal("Each success")
	}
	seen := 0
	r = Enumerator.New(1, 2, 3).Each(func(item interface{}) Result.Interface {
		seen++
		if item.(int) == 2 {
			return boom()
		}
		return Result.New()
	})
	if !r.HasError() || seen != 2 {
		t.Fatal("Each short-circuit")
	}
}

func TestFirst(t *testing.T) {
	if Enumerator.New(7, 8).First().Payload() != 7 {
		t.Fatal("First non-empty")
	}
	if !Enumerator.New().First().HasError() {
		t.Fatal("First empty should error")
	}
}

func TestReduce(t *testing.T) {
	r := Enumerator.New(1, 2, 3, 4).Reduce(0, func(acc, item interface{}) Result.Interface {
		return ok(acc.(int) + item.(int))
	})
	if r.HasError() || r.Payload() != 10 {
		t.Fatal("Reduce sum")
	}
	r = Enumerator.New(1, 2, 3).Reduce(0, func(acc, item interface{}) Result.Interface {
		if item.(int) == 2 {
			return boom()
		}
		return ok(acc.(int) + item.(int))
	})
	if !r.HasError() {
		t.Fatal("Reduce short-circuit")
	}
}

func TestIsNullRealValue(t *testing.T) {
	if Enumerator.New(1).IsNull() {
		t.Fatal("real enumerator should not be null")
	}
}

func TestNullObject(t *testing.T) {
	n := Enumerator.Null()
	if !n.IsNull() {
		t.Fatal("Null().IsNull()")
	}
	// Lazy transforms return the null enumerator.
	if !n.Map(func(x interface{}) interface{} { return x }).IsNull() {
		t.Fatal("null Map")
	}
	if !n.Filter(func(x interface{}) bool { return true }).IsNull() {
		t.Fatal("null Filter")
	}
	if !n.Take(3).IsNull() {
		t.Fatal("null Take")
	}
	if n.ToArray().Len() != 0 {
		t.Fatal("null ToArray")
	}
	if n.Each(func(item interface{}) Result.Interface { return boom() }).HasError() {
		t.Fatal("null Each should be a no-op success")
	}
	if !n.First().HasError() {
		t.Fatal("null First should error")
	}
	if r := n.Reduce(42, func(acc, item interface{}) Result.Interface { return boom() }); r.HasError() || r.Payload() != 42 {
		t.Fatal("null Reduce should return the seed")
	}
}
