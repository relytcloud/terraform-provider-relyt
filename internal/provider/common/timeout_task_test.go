package common

import (
	"errors"
	"testing"
	"time"
)

func TestTimeOutTaskStopsOnTerminalError(t *testing.T) {
	want := errors.New("provisioning failed")
	calls := 0

	start := time.Now()
	_, err := TimeOutTask(60, 5, func() (any, error) {
		calls++
		return nil, Terminal(want)
	})
	elapsed := time.Since(start)

	if calls != 1 {
		t.Errorf("task ran %d times, want 1 — a terminal error must not be retried", calls)
	}
	if !errors.Is(err, want) {
		t.Errorf("got error %v, want the wrapped %v", err, want)
	}
	// One 5s sleep would already blow this; the point is that it returned
	// without waiting out the interval, let alone the 60s timeout.
	if elapsed > time.Second {
		t.Errorf("took %v, want an immediate return", elapsed)
	}
}

func TestTimeOutTaskRetriesOrdinaryError(t *testing.T) {
	calls := 0
	// Interval 1s, timeout 3s: an ordinary error keeps the loop going, so this
	// must poll more than once and then give up with "timeout".
	_, err := TimeOutTask(3, 1, func() (any, error) {
		calls++
		return nil, errors.New("not ready")
	})

	if calls < 2 {
		t.Errorf("task ran %d times, want at least 2 — ordinary errors are retried", calls)
	}
	if err == nil || err.Error() != "timeout" {
		t.Errorf("got error %v, want timeout", err)
	}
}
