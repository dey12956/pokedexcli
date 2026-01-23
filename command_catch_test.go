package main

import (
	"math"
	"testing"
)

const floatTolerance = 1e-9

func TestCatchProbBounds(t *testing.T) {
	if got := catchProb(36); math.Abs(got-0.79) > floatTolerance {
		t.Fatalf("expected 0.79 at min XP, got %v", got)
	}
	if got := catchProb(255); math.Abs(got-0.30) > floatTolerance {
		t.Fatalf("expected 0.30 at kink XP, got %v", got)
	}
	if got := catchProb(608); math.Abs(got-0.08) > floatTolerance {
		t.Fatalf("expected 0.08 at max XP, got %v", got)
	}
}

func TestCatchProbMonotonic(t *testing.T) {
	low := catchProb(100)
	high := catchProb(500)
	if low <= high {
		t.Fatalf("expected lower XP to have higher catch chance: low=%v high=%v", low, high)
	}
}
