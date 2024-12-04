package mathutils

import (
	"math/rand/v2"
	"testing"
)

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(2, 3)
	if result != 6 {
		t.Errorf("Expected 6, got %d", result)
	}
}

func Test_AddFlaky(t *testing.T) {
	if rand.Float32() < 0.7 {
		t.Fatal("Flaky test failed!")
	}
}
