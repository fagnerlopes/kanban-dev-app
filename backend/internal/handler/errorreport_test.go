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

func TestCountTaskOpIsSafeWithoutSentry(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("countTaskOp panicked without an initialized SDK: %v", p)
		}
	}()
	countTaskOp(httptest.NewRequest(http.MethodPost, "/api/tasks", nil), "created")
}

func TestInitSentryIsNoOpWithoutDSN(t *testing.T) {
	if err := InitSentry(config.Config{AppEnv: "local", SentryTracesSampleRate: 1}); err != nil {
		t.Fatalf("InitSentry without a DSN returned %v, want nil", err)
	}
}
