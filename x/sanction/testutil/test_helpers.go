package testutil

import (
	"fmt"
	"runtime/debug"
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

// didPanic safely executes the provided function and returns info about any panic it might have encountered.
func didPanic(f assert.PanicTestFunc) (didPanic bool, message interface{}, stack string) {
	didPanic = true

	defer func() {
		message = recover()
		if didPanic {
			stack = string(debug.Stack())
		}
	}()

	f()
	didPanic = false

	return
}

// AssertPanicsWithMessage asserts that the code inside the specified PanicTestFunc panics, and that
// the recovered panic message equals the expected panic message.
//
//   AssertPanicsWithMessage(t, "crazy error", func(){ GoCrazy() })
//
// PanicsWithValue requires a specific interface{} value to be provided, which can be problematic.
// PanicsWithError requires that the panic value is an error.
// This one uses fmt.Sprintf("%v", panicValue) to convert the panic recovery value to a string to test against.
func AssertPanicsWithMessage(t *testing.T, expected string, f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	t.Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	if !funcDidPanic {
		return assert.Fail(t, fmt.Sprintf("func %#v should panic\n\tPanic value:\t%#v", f, panicValue), msgAndArgs...)
	}
	panicMsg := fmt.Sprintf("%v", panicValue)
	if assert.Equal(t, expected, panicMsg, "panic message") {
		return true
	}
	return assert.Fail(t, fmt.Sprintf("func %#v should panic with value:\t%#v\n\tPanic value:\t%#v\n\tPanic stack:\t%s",
		f, expected, panicValue, panickedStack), msgAndArgs...)
}

// RequirePanicsWithMessage asserts that the code inside the specified PanicTestFunc panics, and that
// the recovered panic message equals the expected panic message.
//
//   RequirePanicsWithMessage(t, "crazy error", func(){ GoCrazy() })
//
// PanicsWithValue requires a specific interface{} value to be provided, which can be problematic.
// PanicsWithError requires that the panic value is an error.
// This one uses fmt.Sprintf("%v", panicValue) to convert the panic recovery value to a string to test against.
func RequirePanicsWithMessage(t *testing.T, expected string, f assert.PanicTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	if AssertPanicsWithMessage(t, expected, f, msgAndArgs...) {
		return
	}
	t.FailNow()
}
