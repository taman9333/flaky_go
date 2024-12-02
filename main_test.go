package main

import (
	"math/rand"
	"testing"
)

func TestStable(t *testing.T) {
	t.Log("This is a stable test")
}

func TestFlaky(t *testing.T) {
	if rand.Intn(2) == 0 {
		t.Fatal("This test is flaky and fails randomly")
	}
}
