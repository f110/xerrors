package xerrors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestError(t *testing.T) {
	e := New("new error")
	require.NotNil(t, e)
	assert.Nil(t, StackTrace(e), "The created error should not contain the stacktrace")
	assert.Nil(t, StackTrace(Newf("new error")), "The created error by Newf should not contain the stacktrace")
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

	wrapped := fmt.Errorf(": %w", e)
	assert.NotNil(t, StackTrace(wrapped))
}

func TestWithMessage(t *testing.T) {
	e := WithMessage(errors.New("raw"), "can not open config file")
	assert.EqualError(t, e, "can not open config file: raw")

	e = WithMessagef(errors.New("raw"), "can not open %s file", "config")
	assert.EqualError(t, e, "can not open config file: raw")
}

func TestMarshalLogArray(t *testing.T) {
	buf := new(bytes.Buffer)
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)

	e := New("new error")
	logger.Info(t.Name(), zap.Array("stacktrace", e.(*Error).stackTrace))

	s := bytes.Split(buf.Bytes(), []byte("\t"))
	assert.Len(t, s, 3)
	data := make(map[string]any)
	err := json.Unmarshal(s[2], &data)
	require.NoError(t, err)
	assert.Contains(t, data, "stacktrace")
	stacktrace, ok := data["stacktrace"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, stacktrace[0], t.Name())
}

func TestIs(t *testing.T) {
	original := New("foo")
	wrap := WithMessage(original, "bar")
	assert.ErrorIs(t, wrap, original)

	assert.True(t, reflect.TypeOf(&Error{}).Comparable(), "Error must be comparable. If the change would make no longer comparable, Is function needs to be implemented.")
}
