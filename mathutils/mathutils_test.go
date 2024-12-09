package mathutils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MathUtilsTestSuite struct {
	suite.Suite
}

func (suite *MathUtilsTestSuite) TestAdd() {
	result := Add(2, 3)
	suite.Equal(5, result, "Expected 5, got %d", result)
}

func (suite *MathUtilsTestSuite) TestMultiply() {
	result := Multiply(2, 3)
	suite.Equal(6, result, "Expected 6, got %d", result)
}

func (suite *MathUtilsTestSuite) Test_AddFlaky() {
	// if rand.Float32() < 0.7 {
	// 	suite.T().Fatal("Flaky test failed!")
	// }
	suite.T().Log("This is a stable test")
}

func TestMathUtilsSuite(t *testing.T) {
	suite.Run(t, new(MathUtilsTestSuite))
}
