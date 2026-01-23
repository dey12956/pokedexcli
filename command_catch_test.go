package main

import "testing"

func TestCatchProbBounds(t *testing.T) {
	if got := catchProb(36); got != 0.95 {
		t.Fatalf("expected 0.95 at min XP, got %v", got)
	}
	if got := catchProb(255); got != 0.40 {
		t.Fatalf("expected 0.40 at kink XP, got %v", got)
	}
	if got := catchProb(608); got != 0.08 {
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
