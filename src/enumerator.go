// Package Enumerator is a lazy sequence composite in the spirit of Ruby's
// Enumerator::Lazy. Transforms (Map/Filter/Take) are lazy — they wrap a
// producer and evaluate nothing until a terminal operation (ToArray/Each/
// First/Reduce) pulls items through. This makes it safe to Take a finite prefix
// of an infinite Generate source.
package Enumerator

import (
	Array "github.com/go-composites/array/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

// Interface is the lazy sequence contract.
type Interface interface {
	// Lazy transforms — return a new Enumerator, evaluate nothing.
	Map(fn func(interface{}) interface{}) Interface
	Filter(pred func(interface{}) bool) Interface
	Take(n int) Interface
	// Terminal operations — force evaluation.
	ToArray() Array.Interface
	Each(fn func(interface{}) Result.Interface) Result.Interface
	First() Result.Interface
	Reduce(seed interface{}, fn func(acc, item interface{}) Result.Interface) Result.Interface
	IsNull() bool
}

// data drives the sequence: each calls yield(item) for every element and stops
// early when yield returns false (that early stop is what makes Take/First
// lazy and lets them terminate an infinite producer).
type data struct {
	each func(yield func(item interface{}) bool)
}

// New builds a finite Enumerator over the given items.
func New(items ...interface{}) Interface {
	return &data{each: func(yield func(interface{}) bool) {
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}}
}

// Generate builds an INFINITE Enumerator: seed, fn(seed), fn(fn(seed)), …
// It only terminates when a downstream consumer (e.g. Take/First) stops it.
func Generate(seed interface{}, fn func(interface{}) interface{}) Interface {
	return &data{each: func(yield func(interface{}) bool) {
		current := seed
		for yield(current) {
			current = fn(current)
		}
	}}
}

func (d *data) Map(fn func(interface{}) interface{}) Interface {
	parent := d.each
	return &data{each: func(yield func(interface{}) bool) {
		parent(func(item interface{}) bool {
			return yield(fn(item))
		})
	}}
}

func (d *data) Filter(pred func(interface{}) bool) Interface {
	parent := d.each
	return &data{each: func(yield func(interface{}) bool) {
		parent(func(item interface{}) bool {
			if pred(item) {
				return yield(item)
			}
			return true
		})
	}}
}

func (d *data) Take(n int) Interface {
	parent := d.each
	return &data{each: func(yield func(interface{}) bool) {
		if n <= 0 {
			return
		}
		count := 0
		parent(func(item interface{}) bool {
			count++
			ok := yield(item)
			return ok && count < n
		})
	}}
}

func (d *data) ToArray() Array.Interface {
	out := Array.New()
	d.each(func(item interface{}) bool {
		out.Push(item)
		return true
	})
	return out
}

func (d *data) Each(fn func(interface{}) Result.Interface) Result.Interface {
	var errResult Result.Interface
	d.each(func(item interface{}) bool {
		if result := fn(item); result.HasError() {
			errResult = result
			return false
		}
		return true
	})
	if errResult != nil {
		return errResult
	}
	return Result.New()
}

func (d *data) First() Result.Interface {
	var first interface{}
	found := false
	d.each(func(item interface{}) bool {
		first = item
		found = true
		return false
	})
	if found {
		return Result.New(Result.WithPayload(first))
	}
	return Result.New(Result.WithError(Error.New("Enumerator.First: empty")))
}

func (d *data) Reduce(
	seed interface{},
	fn func(acc, item interface{}) Result.Interface,
) Result.Interface {
	acc := seed
	var errResult Result.Interface
	d.each(func(item interface{}) bool {
		result := fn(acc, item)
		if result.HasError() {
			errResult = result
			return false
		}
		acc = result.Payload()
		return true
	})
	if errResult != nil {
		return errResult
	}
	return Result.New(Result.WithPayload(acc))
}

func (d *data) IsNull() bool {
	return false
}

// null is the Null-Object Enumerator: an empty lazy sequence that honours the
// whole Interface without ever being nil. Transforms return the null
// enumerator; terminals yield nothing.
type null struct{}

// Null returns the Null-Object Enumerator.
func Null() Interface {
	return &null{}
}

func (n *null) Map(fn func(interface{}) interface{}) Interface { return n }
func (n *null) Filter(pred func(interface{}) bool) Interface   { return n }
func (n *null) Take(count int) Interface                       { return n }
func (n *null) ToArray() Array.Interface                       { return Array.New() }
func (n *null) Each(fn func(interface{}) Result.Interface) Result.Interface {
	return Result.New()
}

func (n *null) First() Result.Interface {
	return Result.New(Result.WithError(Error.New("Enumerator.First: empty")))
}

func (n *null) Reduce(
	seed interface{},
	fn func(acc, item interface{}) Result.Interface,
) Result.Interface {
	return Result.New(Result.WithPayload(seed))
}

func (n *null) IsNull() bool {
	return true
}
