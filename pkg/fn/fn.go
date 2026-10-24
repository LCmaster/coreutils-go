// Package fn provides generic functional programming primitives and higher-order
// combinators used across core utilities for declarative data transformation and filtering.
package fn

// AnyMatch reports whether at least one element in the provided slice satisfies
// the given predicate function.
//
// If the items slice is empty or nil, AnyMatch immediately returns false without
// invoking the predicate. The function short-circuits on the first matching element.
func AnyMatch[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns a composite predicate that evaluates to true only if every provided
// predicate returns true for a given input value.
//
// Predicates are evaluated in order from left to right, short-circuiting on the first
// predicate that returns false. If no predicates are supplied, the returned function
// evaluates to true for all inputs (the identity value for logical conjunction).
func All[T any](predicates ...func(T) bool) func(T) bool {
	return func(val T) bool {
		for _, pred := range predicates {
			if !pred(val) {
				return false
			}
		}
		return true
	}
}

// Pipe composes a sequence of unary transformer functions into a single pipeline
// function applied from left to right (function composition).
//
// The output of each transformer function is passed as the input to the next:
//
//	Pipe(f, g, h)(x) == h(g(f(x)))
//
// If no transformers are supplied, Pipe returns an identity function that returns
// the input unchanged.
func Pipe[T any](transforms ...func(T) T) func(T) T {
	return func(val T) T {
		for _, transform := range transforms {
			val = transform(val)
		}
		return val
	}
}
