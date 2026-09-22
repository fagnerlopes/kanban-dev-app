package handler

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
)

// NewAppLogger builds the process-wide logger: plain text to stderr, plus a tee
// into Sentry's structured logs from Warn up.
//
// The base handler is constructed here, explicitly, and that is the whole
// point. Wrapping slog.Default().Handler() instead looks equivalent and
// deadlocks the process on its first log line: slog's built-in default handler
// writes through the standard log package, and slog.SetDefault routes that
// package back into the slog default — which would be this handler. The two
// call each other until the log package's mutex locks against itself, with no
// panic and no message. Never wrap the default handler and then SetDefault it.
func NewAppLogger(dsn string) *slog.Logger {
	base := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(NewSentrySlogHandler(base, slog.LevelWarn, dsn))
}

// NewSentrySlogHandler tees slog records: they keep going to base (so
// `docker logs` still shows everything) and, from minLevel up, they are also
// emitted to Sentry's structured logs.
//
// A bridge rather than a new logging call at each site: the app already logs
// with slog everywhere, so this turns the existing logs into Sentry logs
// without touching a single caller. The Sentry Go SDK has no "enable logs"
// option -- using sentry.NewLogger is what switches the feature on.
//
// Returns base unchanged when Sentry is not configured, so local runs are
// untouched.
func NewSentrySlogHandler(base slog.Handler, minLevel slog.Level, dsn string) slog.Handler {
	if dsn == "" {
		return base
	}
	return &sentrySlogHandler{base: base, minLevel: minLevel}
}

type sentrySlogHandler struct {
	base     slog.Handler
	minLevel slog.Level
	attrs    []slog.Attr
	groups   []string
}

func (h *sentrySlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *sentrySlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sentrySlogHandler{
		base:     h.base.WithAttrs(attrs),
		minLevel: h.minLevel,
		attrs:    append(append([]slog.Attr{}, h.attrs...), attrs...),
		groups:   h.groups,
	}
}

func (h *sentrySlogHandler) WithGroup(name string) slog.Handler {
	return &sentrySlogHandler{
		base:     h.base.WithGroup(name),
		minLevel: h.minLevel,
		attrs:    h.attrs,
		groups:   append(append([]string{}, h.groups...), name),
	}
}

func (h *sentrySlogHandler) Handle(ctx context.Context, r slog.Record) error {
	// The local log is the one that must never be lost, so it goes first and
	// its error is what we return.
	err := h.base.Handle(ctx, r)
	if r.Level < h.minLevel {
		return err
	}

	entry := logEntryFor(sentry.NewLogger(ctx), r.Level)
	if entry == nil {
		return err
	}
	for _, a := range h.attrs {
		entry = withAttr(entry, h.groups, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		entry = withAttr(entry, h.groups, a)
		return true
	})
	entry.Emit(r.Message)
	return err
}

// levelName maps a slog level onto Sentry's. Split out from logEntryFor so the
// mapping can be asserted without an initialized SDK.
func levelName(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "error"
	case level >= slog.LevelWarn:
		return "warn"
	case level >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

func logEntryFor(l sentry.Logger, level slog.Level) sentry.LogEntry {
	switch levelName(level) {
	case "error":
		return l.Error()
	case "warn":
		return l.Warn()
	case "info":
		return l.Info()
	default:
		return l.Debug()
	}
}

// withAttr maps a slog attribute onto the Sentry entry, keeping the native
// type where there is one so the value stays filterable in Sentry instead of
// arriving as a string.
func withAttr(e sentry.LogEntry, groups []string, a slog.Attr) sentry.LogEntry {
	key := a.Key
	for i := len(groups) - 1; i >= 0; i-- {
		key = groups[i] + "." + key
	}

	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindBool:
		return e.Bool(key, v.Bool())
	case slog.KindInt64:
		return e.Int64(key, v.Int64())
	case slog.KindUint64:
		// No Uint64 on the entry API; Int64 keeps it numeric and the values
		// this app logs (ports, counts) are nowhere near overflowing.
		return e.Int64(key, int64(v.Uint64()))
	case slog.KindFloat64:
		return e.Float64(key, v.Float64())
	case slog.KindString:
		return e.String(key, v.String())
	case slog.KindGroup:
		for _, ga := range v.Group() {
			e = withAttr(e, append(groups, a.Key), ga)
		}
		return e
	default:
		// Durations, times, errors and anything custom.
		return e.String(key, fmt.Sprint(v.Any()))
	}
}
