// Package log implements an opinionated structured logger for Cloud Run
// - No dependencies
// - Go's standard JSON is fast enough
// - No formatting functions as you should use key values for this
package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type ContextKey string

const (
	// TraceHeader is the key of the Google Cloud trace header
	TraceHeader = "X-Cloud-Trace-Context"

	// CloudTraceContextKey is the context.Context key used for the X-Cloud-Trace-Context request header.
	CloudTraceContextKey ContextKey = TraceHeader
)

type Logger struct {
	projectID string
	logger    *log.Logger
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func KV(key string, value interface{}) KeyValue {
	return KeyValue{key, value}
}

// NewJSONLogger generates a structured logger
// redirectStdLog redirects output from the standard library's package-global logger to the this logger at Info level.
func NewJSONLogger(out io.Writer, projectID string, redirectStdLog bool) *Logger {
	logger := log.New(out, "", 0)

	l := &Logger{projectID, logger}
	if redirectStdLog {
		log.SetFlags(0)
		log.SetPrefix("")
		log.SetOutput(l)
	}

	return l
}

// Info logs message with severity INFO
func (l *Logger) Info(ctx context.Context, msg string, keysValues ...KeyValue) {
	l.writeLog(ctx, "INFO", msg, keysValues)
}

// Warn logs message with severity WARNING
func (l *Logger) Warn(ctx context.Context, msg string, keysValues ...KeyValue) {
	l.writeLog(ctx, "WARNING", msg, keysValues)
}

// Error logs message with severity ERROR and includes stacktrace
func (l *Logger) Error(ctx context.Context, err error, keysValues ...KeyValue) {
	l.writeLog(ctx, "ERROR", toStacktrace(err), keysValues)
}

// Fatal logs message with severity CRITICAL and includes stacktrace, then calls os.Exit(1)
func (l *Logger) Fatal(ctx context.Context, err error, keysValues ...KeyValue) {
	l.writeLog(ctx, "CRITICAL", toStacktrace(err), keysValues)
	os.Exit(1)
}

// Write implements io.Writer which logs at Info level (used for redirecting StdLog)
func (l *Logger) Write(p []byte) (int, error) {
	ctx := context.Background()

	p = bytes.TrimSpace(p)
	l.Info(ctx, string(p))
	return len(p), nil
}

// Error reporting stacktrace must be the return value of runtime.Stack().
func toStacktrace(err error) string {
	//limit the stack trace to 16k.
	var buf [16 * 1024]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	return err.Error() + "\n" + string(stack)
}

func (l *Logger) writeLog(ctx context.Context, level string, msg string, keysValues []KeyValue) {
	m := []KeyValue{
		{"time", time.Now().UTC().Format(time.RFC3339Nano)},
		{"severity", level},
		{"message", msg},
	}
	m = append(m, keysValues...)

	// Add trace if available in context
	if l.projectID != "" {
		cloudTraceHeader, ok := ctx.Value(CloudTraceContextKey).(string)
		if ok && cloudTraceHeader != "" {
			// Format: X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE
			traceParts := strings.Split(cloudTraceHeader, "/")

			if len(traceParts) > 0 && traceParts[0] != "" {
				trace := fmt.Sprintf("projects/%s/traces/%s", l.projectID, traceParts[0])
				m = append(m, KeyValue{"logging.googleapis.com/trace", trace})
			}
		}
	}

	out, err := l.marshalJSON(m)
	if err != nil {
		l.logger.Printf("unable to marshal json log: %v", err)
		return
	}

	//nolint:errcheck
	l.logger.Output(0, out)
}

func (l *Logger) marshalJSON(keysValues []KeyValue) (string, error) {
	var buf strings.Builder

	buf.WriteByte('{')
	for i, kv := range keysValues {
		if i > 0 {
			buf.WriteByte(',')
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return "", err
		}
		buf.Write(key)
		buf.WriteByte(':')
		// marshal value
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return "", err
		}
		buf.Write(val)
	}

	buf.WriteByte('}')

	return buf.String(), nil
}

// Tracer is a http handler for adding the X-Cloud-Trace-Context header to the context.Context
// when enabled traces are automatically added to log statements
func (l *Logger) Tracer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), CloudTraceContextKey, r.Header.Get(TraceHeader)))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
