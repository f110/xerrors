package xerrors

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	e := New("new error")
	require.NotNil(t, e)
	_, ok := e.(interface {
		Unwrap() error
	})
	assert.True(t, ok)
	assert.EqualError(t, e, "new error")
	assert.Implements(t, (*fmt.Formatter)(nil), e)

	e = WithStack(errors.New("root cause"))
	assert.EqualError(t, e, "root cause")
	v, _ := e.(interface {
		Unwrap() error
	})
	assert.NotNil(t, v.Unwrap())
	assert.EqualError(t, v.Unwrap(), "root cause")
}

func TestStackTrace(t *testing.T) {
	e := errors.New("bare")
	assert.Nil(t, StackTrace(e))

	e = WithStack(e)
	frames := StackTrace(e)
	_, filename, _, _ := runtime.Caller(1)
	assert.Contains(t, frames.String(), filename)
}

func TestWithMessage(t *testing.T) {
	e := WithMessage(errors.New("raw"), "can not open config file")
	assert.EqualError(t, e, "can not open config file: raw")

	e = WithMessagef(errors.New("raw"), "can not open %s file", "config")
	assert.EqualError(t, e, "can not open config file: raw")
}
