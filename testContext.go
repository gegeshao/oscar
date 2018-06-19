package oscar

import (
	"fmt"
	"regexp"
	"time"
)

// TestContext is nested structure, that holds test invocation context
type TestContext struct {
	Parent *TestContext

	Vars  map[string]string
	Error error

	CountAssertSuccess  int
	CountRemoteRequests int

	elapsedHTTP, elapsedSleep time.Duration

	OnFinish func(*TestContext) error
	OnEvent  func(interface{})

	startedAt, finishedAt time.Time
}

// Get returns variable value from vars map
func (t *TestContext) Get(key string) string {
	if len(t.Vars) > 0 {
		if v, ok := t.Vars[key]; ok {
			return v
		}
	}
	if t.Parent != nil {
		return t.Parent.Get(key)
	}

	return ""
}

// Set assigns new variable value
func (t *TestContext) Set(key, value string) {
	t.Trace(`Setting "%s" := "%s"`, key, value)
	if len(t.Vars) == 0 {
		t.Vars = map[string]string{}
	}

	t.Vars[key] = value
}

var iregex = regexp.MustCompile(`\${([\w.-]+)}`)

// Interpolate replaces all placeholders in provided string using vars from test case or
// global runner
func (t *TestContext) Interpolate(value string) string {
	return iregex.ReplaceAllStringFunc(value, func(i string) string {
		m := iregex.FindStringSubmatch(i)
		return t.Get(m[1])
	})
}

// Elapsed returns elapsed time
func (t *TestContext) Elapsed() (total time.Duration, http time.Duration, sleep time.Duration) {
	total = t.finishedAt.Sub(t.startedAt)
	http = t.elapsedHTTP
	sleep = t.elapsedSleep
	return
}

// Emit publishes new event into nested test context
func (t *TestContext) Emit(event interface{}) {
	if s, ok := event.(StartEvent); ok {
		if t.startedAt.IsZero() {
			t.startedAt = s.Time
		}
	} else if s, ok := event.(FinishEvent); ok {
		t.finishedAt = s.Time
	} else if _, ok := event.(AssertionSuccess); ok {
		t.CountAssertSuccess++
	} else if a, ok := event.(AssertionFailure); ok {
		if t.Error == nil {
			t.Error = a
		}
	} else if r, ok := event.(RemoteRequestEvent); ok {
		t.CountRemoteRequests++
		t.elapsedHTTP += r.Elapsed
	} else if s, ok := event.(SleepEvent); ok {
		t.elapsedSleep += time.Duration(s)
	}

	if t.OnEvent != nil {
		t.OnEvent(event)
	}
	if s, ok := event.(FinishEvent); ok && t.OnFinish != nil {
		if t, ok := s.Owner.(*TestCase); ok {
			t.OnFinish(t.TestContext)
		}
	}

	if t.Parent != nil {
		t.Parent.Emit(event)
	}
}

// Trace emits tracing event
func (t *TestContext) Trace(pattern string, args ...interface{}) {
	t.Emit(TraceEvent(fmt.Sprintf(pattern, args...)))
}
