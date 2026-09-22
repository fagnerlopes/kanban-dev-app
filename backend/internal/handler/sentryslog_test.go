package handler

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// The regression that cost an afternoon: building the app logger by wrapping
// slog.Default().Handler() and then installing it with slog.SetDefault makes
// slog and the standard log package call each other until the log package's
// mutex locks against itself. The process hangs on its FIRST log line, with no
// panic and no message — it simply stops, still "running".
//
// The timeout is the assertion: if the cycle comes back, this test hangs
// instead of failing, and a hang is what the guard is for.
func TestAppLoggerDoesNotDeadlockAfterSetDefault(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	slog.SetDefault(NewAppLogger("https://publickey@o1.ingest.sentry.io/2"))

	done := make(chan struct{})
	go func() {
		defer close(done)
		slog.Warn("primeira linha apos SetDefault", "origem", "teste")
		slog.Error("segunda linha", "origem", "teste")
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("logging deadlocked after slog.SetDefault — the log/slog cycle is back")
	}
}

// Without a DSN the bridge must disappear entirely — local runs should not pay
// for, or be changed by, machinery that has nowhere to send anything.
func TestSentrySlogHandlerIsTransparentWithoutDSN(t *testing.T) {
	base := slog.NewTextHandler(&bytes.Buffer{}, nil)

	if got := NewSentrySlogHandler(base, slog.LevelWarn, ""); got != base {
		t.Errorf("handler = %T, want the base handler returned unchanged", got)
	}
}

// The local log is the contract that must never break: whatever happens on the
// Sentry side, `docker logs` still has to show every record, at every level.
func TestSentrySlogHandlerAlwaysWritesLocally(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log := slog.New(NewSentrySlogHandler(base, slog.LevelWarn, "https://k@o1.ingest.sentry.io/2"))

	log.Debug("abaixo do corte", "n", 1)
	log.Error("acima do corte", "n", 2)

	out := buf.String()
	for _, want := range []string{"abaixo do corte", "acima do corte"} {
		if !strings.Contains(out, want) {
			t.Errorf("local output lost %q; got: %s", want, out)
		}
	}
}

// Attributes added with With() must survive to the record, and groups must
// prefix the key — otherwise two different "id" fields collide in Sentry.
func TestSentrySlogHandlerKeepsAttrsAndGroups(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log := slog.New(NewSentrySlogHandler(base, slog.LevelWarn, "https://k@o1.ingest.sentry.io/2")).
		With("service", "kanban").
		WithGroup("req")

	log.Error("falhou", "path", "/api/board")

	out := buf.String()
	for _, want := range []string{"service=kanban", "req.path=/api/board"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got: %s", want, out)
		}
	}
}

// Emitting to Sentry with no SDK initialized must not panic — the app has to
// survive a missing or broken DSN, and this bridge runs on every warning.
func TestSentrySlogHandlerSurvivesUninitializedSDK(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	h := NewSentrySlogHandler(base, slog.LevelWarn, "https://k@o1.ingest.sentry.io/2")

	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("bridge panicked without an initialized SDK: %v", p)
		}
	}()

	rec := slog.NewRecord(time.Now(), slog.LevelError, "boom", 0)
	rec.AddAttrs(
		slog.Bool("ok", false),
		slog.Int64("count", 7),
		slog.Float64("ratio", 0.5),
		slog.String("name", "x"),
		slog.Duration("took", 0),
		slog.Group("db", slog.String("host", "db")),
	)
	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle returned %v", err)
	}
}

// Level mapping: anything at or above the cut must reach Sentry's entry API,
// and the severity must not collapse to a single level.
func TestLogEntryForMapsLevels(t *testing.T) {
	for _, tc := range []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelDebug, "debug"},
		{slog.LevelInfo, "info"},
		{slog.LevelWarn, "warn"},
		{slog.LevelError, "error"},
	} {
		if got := levelName(tc.level); got != tc.want {
			t.Errorf("level %v mapped to %q, want %q", tc.level, got, tc.want)
		}
	}
}
