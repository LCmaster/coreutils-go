package fn

import "testing"

func TestAnyMatch(t *testing.T) {
	numbers := []int{1, 3, 5, 8, 9}

	isEven := func(n int) bool { return n%2 == 0 }
	isNegative := func(n int) bool { return n < 0 }

	if !AnyMatch(numbers, isEven) {
		t.Errorf("expected AnyMatch with isEven to be true")
	}

	if AnyMatch(numbers, isNegative) {
		t.Errorf("expected AnyMatch with isNegative to be false")
	}

	if AnyMatch(nil, isEven) {
		t.Errorf("expected AnyMatch on empty slice to be false")
	}
}

func TestAll(t *testing.T) {
	isPositive := func(n int) bool { return n > 0 }
	isEven := func(n int) bool { return n%2 == 0 }
	isLessThanTen := func(n int) bool { return n < 10 }

	allPred := All(isPositive, isEven, isLessThanTen)

	if !allPred(4) {
		t.Errorf("expected allPred(4) to be true")
	}

	if allPred(-2) {
		t.Errorf("expected allPred(-2) to be false (not positive)")
	}

	if allPred(5) {
		t.Errorf("expected allPred(5) to be false (not even)")
	}

	if allPred(12) {
		t.Errorf("expected allPred(12) to be false (not less than 10)")
	}
}

func TestPipe(t *testing.T) {
	addTwo := func(x int) int { return x + 2 }
	multiplyThree := func(x int) int { return x * 3 }
	subtractOne := func(x int) int { return x - 1 }

	pipeline := Pipe(addTwo, multiplyThree, subtractOne)
	// (5 + 2) * 3 - 1 = 20
	got := pipeline(5)
	if got != 20 {
		t.Errorf("Pipe(5) = %d, want 20", got)
	}

	emptyPipeline := Pipe[int]()
	if emptyPipeline(42) != 42 {
		t.Errorf("Empty Pipe(42) = %d, want 42", emptyPipeline(42))
	}
}
