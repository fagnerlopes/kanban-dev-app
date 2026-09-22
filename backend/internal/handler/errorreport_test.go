package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"kanban-dev-app/backend/internal/config"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
)

// spyTransport captures what the SDK would have sent, so these tests assert on
// real SDK behaviour instead of trusting that a call was made.
type spyTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (s *spyTransport) Configure(sentry.ClientOptions) {}
func (s *spyTransport) SendEvent(e *sentry.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}
func (s *spyTransport) Flush(time.Duration) bool              { return true }
func (s *spyTransport) FlushWithContext(context.Context) bool { return true }
func (s *spyTransport) Close()                                {}

func (s *spyTransport) captured() []*sentry.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*sentry.Event{}, s.events...)
}

// withSpySentry initializes the SDK against a spy transport for the duration of
// one test, restoring the previous hub afterwards.
func withSpySentry(t *testing.T) *spyTransport {
	t.Helper()
	spy := &spyTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://publickey@o1.ingest.sentry.io/2",
		Transport: spy,
	})
	if err != nil {
		t.Fatalf("sentry client: %v", err)
	}
	previous := sentry.CurrentHub().Client()
	sentry.CurrentHub().BindClient(client)
	t.Cleanup(func() { sentry.CurrentHub().BindClient(previous) })
	return spy
}

// The gap this closes: before, a failing query answered 500 and told Sentry
// nothing, because only panics were reported. Handing that error to an agent
// was impossible — there was no record of it anywhere but the container log.
func TestWriteErrReportsServerErrors(t *testing.T) {
	spy := withSpySentry(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/board", nil)

	writeErr(rec, req, errors.New("relation \"tasks\" does not exist"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	events := spy.captured()
	if len(events) != 1 {
		t.Fatalf("captured %d events, want exactly 1", len(events))
	}
	if got := events[0].Exception; len(got) == 0 || got[0].Value == "" {
		t.Fatalf("event carries no exception: %+v", events[0])
	}
}

// A missing row is an ordinary 404, not an incident. Reporting it would bury
// the real errors under noise from every stale link.
func TestWriteErrDoesNotReportNotFound(t *testing.T) {
	spy := withSpySentry(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/999", nil)

	writeErr(rec, req, pgx.ErrNoRows)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if n := len(spy.captured()); n != 0 {
		t.Fatalf("captured %d events for a 404, want 0", n)
	}
}

// Metrics run on every successful board operation, including in local dev with
// no DSN. A nil meter there would crash the request it was measuring.
func TestCountTaskOpIsSafeWithoutSentry(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("countTaskOp panicked without an initialized SDK: %v", p)
		}
	}()
	countTaskOp(httptest.NewRequest(http.MethodPost, "/api/tasks", nil), "created")
}

// Tracing must stay off unless a DSN is configured, and InitSentry must never
// be the thing that stops the app from booting.
func TestInitSentryIsNoOpWithoutDSN(t *testing.T) {
	if err := InitSentry(config.Config{AppEnv: "local", SentryTracesSampleRate: 1}); err != nil {
		t.Fatalf("InitSentry without a DSN returned %v, want nil", err)
	}
}
