package tests

// Basic imports
import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConchTestSuite struct {
	suite.Suite
}

func (suite *ConchTestSuite) SetupTest() {
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ConchTestSuite))
}
