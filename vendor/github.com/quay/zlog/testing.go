package zlog

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/baggage"
)

const (
	testNameKey = "zlog.testname"
)

var (
	setup sync.Once
	sink  logsink
)

// Test configures and wires up the global logger for testing.
//
// Once called, log messages that do not use a Context returned by a call to
// Test will cause a panic.
//
// Passing a nil Context will return a Context derived from context.Background.
func Test(ctx context.Context, t testing.TB) context.Context {
	t.Helper()
	setup.Do(sink.Setup)
	t.Cleanup(func() {
		sink.Remove(t)
	})
	sink.Create(t)
	if ctx == nil {
		ctx = context.Background()
	}
	m, err := baggage.NewMember(testNameKey, t.Name())
	if err != nil {
		t.Fatal(err)
	}
	b, err := baggage.FromContext(ctx).SetMember(m)
	if err != nil {
		t.Fatal(err)
	}
	return baggage.ContextWithBaggage(ctx, b)
}

// Logsink holds the files and does the routing for log messages.
type logsink struct {
	mu sync.RWMutex
	ts map[string]testing.TB
}

// Setup configures the logsink and configures the logger.
func (s *logsink) Setup() {
	s.ts = make(map[string]testing.TB)

	// Set up caller information be default, because the testing package's line
	// information will be incorrect.
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	l := zerolog.New(s).With().Caller().Logger()
	Set(&l)
}

// Create initializes a new log stream.
func (s *logsink) Create(t testing.TB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ts[t.Name()] = t
}

// Remove tears down a log stream.
func (s *logsink) Remove(t testing.TB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ts, t.Name())
}

// Write routes writes to the correct stream.
func (s *logsink) Write(b []byte) (int, error) {
	var ev ev
	if err := json.Unmarshal(b, &ev); err != nil {
		return -1, err
	}
	l := len(b)
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.ts[ev.Name]
	if !ok {
		panic(fmt.Sprintf("log write to unknown test %q:\n%s", ev.Name, string(b)))
	}
	t.Log(string(b[:l-1]))
	return l, nil
}

// Ev is used to pull the test name out of the zerolog Event.
type ev struct {
	Name string `json:"zlog.testname"`
}
