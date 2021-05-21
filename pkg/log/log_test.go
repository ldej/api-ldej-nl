package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdLog "log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ldej/api-ldej-nl/pkg/log"
)

type LogLine struct {
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"`
	Key      string    `json:"key"`
	Trace    string    `json:"logging.googleapis.com/trace,omitempty"`
}

func TestStdLogger(t *testing.T) {
	flags := stdLog.Flags()
	prefix := stdLog.Prefix()
	out := stdLog.Writer()

	msg := "infoMsg"

	var b bytes.Buffer
	log.NewJSONLogger(&b, "my-project", true)

	stdLog.Println(msg)

	var result LogLine
	err := json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, "INFO", result.Severity)

	// Reset standard logger
	stdLog.SetFlags(flags)
	stdLog.SetPrefix(prefix)
	stdLog.SetOutput(out)
}

func TestInfo(t *testing.T) {
	ctx := context.Background()
	msg := "infoMsg"
	kv := log.KV("key", "value")

	var b bytes.Buffer
	l := log.NewJSONLogger(&b, "my-project", false)

	l.Info(ctx, msg, kv)

	var result LogLine
	err := json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.LessOrEqual(t, result.Time.UnixNano(), time.Now().UTC().UnixNano())
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, "INFO", result.Severity)
	assert.Equal(t, kv.Value, result.Key)
}

func TestWarn(t *testing.T) {
	ctx := context.Background()
	msg := "warningMsg"

	var b bytes.Buffer
	l := log.NewJSONLogger(&b, "my-project", false)

	l.Warn(ctx, msg)

	var result LogLine
	err := json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, "WARNING", result.Severity)
}

func TestError(t *testing.T) {
	ctx := context.Background()
	errIn := errors.New("errorMsg")

	var b bytes.Buffer
	l := log.NewJSONLogger(&b, "my-project", false)

	l.Error(ctx, errIn)

	var result LogLine
	err := json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.Contains(t, result.Message, fmt.Sprintf("%s\ngoroutine", errIn))
	assert.Equal(t, "ERROR", result.Severity)
}

func TestWrite(t *testing.T) {
	msg := "msg"

	var b bytes.Buffer
	l := log.NewJSONLogger(&b, "my-project", false)

	_, err := l.Write([]byte(msg))
	assert.NoError(t, err)

	var result LogLine
	err = json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, "INFO", result.Severity)
}

func TestTrace(t *testing.T) {
	ctx := context.Background()
	msg := "infoMsg"

	var b bytes.Buffer
	l := log.NewJSONLogger(&b, "my-project", false)

	ctxWithTrace := context.WithValue(ctx, log.CloudTraceContextKey, "58baa99db36bc04802c1519f8769901a/0;o=1")
	l.Info(ctxWithTrace, msg)

	var result LogLine
	err := json.Unmarshal(b.Bytes(), &result)

	assert.NoError(t, err)
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, "INFO", result.Severity)
	assert.Equal(t, "projects/my-project/traces/58baa99db36bc04802c1519f8769901a", result.Trace)
}

func TestTracer(t *testing.T) {
	traceValue := "58baa99db36bc04802c1519f8769901a/0;o=1"

	l := log.NewJSONLogger(os.Stderr, "my-project", false)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cloudTraceHeader, ok := r.Context().Value(log.CloudTraceContextKey).(string)
		assert.True(t, ok)
		assert.Equal(t, traceValue, cloudTraceHeader)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.action.com", nil)
	req.Header.Set(log.TraceHeader, traceValue)

	l.Tracer(testHandler).ServeHTTP(nil, req)
}
