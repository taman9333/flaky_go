package main

import (
	"math/rand/v2"
	"testing"
)

func TestStable(t *testing.T) {
	t.Log("This is a stable test")
}

// func TestFlaky(t *testing.T) {
// 	if rand.Intn(2) == 0 {
// 		t.Fatal("This test is flaky and fails randomly")
// 	}
// }

// handle if all test cases failed
func Test_Flaky2(t *testing.T) {
	if rand.Float32() < 0.8 {
		t.Fatal("Flaky test failed!")
	}
}
