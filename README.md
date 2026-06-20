<p align="center"><img src="https://raw.githubusercontent.com/go-composites/brand/main/social/go-composites.png" alt="go-composites/enumerator" width="720"></p>

# enumerator

[![ci](https://github.com/go-composites/enumerator/actions/workflows/ci.yml/badge.svg)](https://github.com/go-composites/enumerator/actions/workflows/ci.yml)

The **lazy sequence** composite of [go-composites](https://github.com/go-composites),
in the spirit of Ruby's `Enumerator::Lazy`. Transforms are *lazy* — they wrap a
producer and evaluate nothing until a terminal operation pulls items through —
so you can take a finite prefix of an **infinite** source without it ever running
away.

## Install

```sh
go get github.com/go-composites/enumerator
```

## API

```go
import Enumerator "github.com/go-composites/enumerator/src"

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

func New(items ...interface{}) Interface                                   // finite source
func Generate(seed interface{}, fn func(interface{}) interface{}) Interface // infinite source
func Null() Interface                                                       // Null-Object
```

| Kind | Method | Behaviour |
| --- | --- | --- |
| lazy | `Map(fn)` | yields `fn(item)` for each upstream item |
| lazy | `Filter(pred)` | yields only items where `pred(item)` is true |
| lazy | `Take(n)` | yields at most `n` items, then **stops the upstream producer** (terminates an infinite source; `n ≤ 0` is empty) |
| terminal | `ToArray()` | materialise into a go-composites [`Array`](https://github.com/go-composites/array) |
| terminal | `Each(fn)` | run `fn` per item; **short-circuit** on the first error `Result` |
| terminal | `First()` | a `Result` with the first item; empty → an error `Result` |
| terminal | `Reduce(seed, fn)` | left fold, short-circuiting on the first error `Result` |

```go
// A finite prefix of an infinite sequence — lazily.
squares := Enumerator.
    Generate(1, func(x interface{}) interface{} { return x.(int) + 1 }).
    Map(func(x interface{}) interface{} { return x.(int) * x.(int) }).
    Take(3).
    ToArray() // [1 4 9]
```

`Null()` is the never-nil empty sequence: transforms return it, `ToArray` is
empty, `First` errors, `Each`/`Reduce` are no-op successes, and `IsNull()` is
`true`.

## Dependencies

`enumerator` depends on [`array`](https://github.com/go-composites/array) (for
`ToArray`), [`result`](https://github.com/go-composites/result) and
[`error`](https://github.com/go-composites/error), and transitively on
[`null`](https://github.com/go-composites/null).

## License

BSD-3-Clause © the go-composites/enumerator authors.
