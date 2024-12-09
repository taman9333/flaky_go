package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite
}

func (suite *MainTestSuite) TestStable() {
	suite.T().Log("This is a stable test")
}

// Simulate flaky test
func (suite *MainTestSuite) Test_Flaky2() {
	// if rand.Float32() < 0.8 {
	suite.T().Fatal("Flaky test failed!")
	// }
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
