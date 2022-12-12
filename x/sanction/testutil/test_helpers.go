package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file contains some functions handy for doing unit tests.

// AssertErrorContents asserts that, if contains is empty, there's no error.
// Otherwise, asserts that there is an error, and that it contains each of the provided strings.
func AssertErrorContents(t *testing.T, theError error, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(contains) == 0 {
		return assert.NoError(t, theError, msgAndArgs)
	}
	rv := assert.Error(t, theError, msgAndArgs...)
	if rv {
		for _, expInErr := range contains {
			rv = assert.ErrorContains(t, theError, expInErr, msgAndArgs...) && rv
		}
	}
	return rv
}
