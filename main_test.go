package main

import (
	"math/rand"
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

func Test_Flaky2(t *testing.T) {
	if rand.Float32() < 0.9 {
		t.Fatal("Flaky test failed!")
	}
}
