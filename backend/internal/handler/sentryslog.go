package handler

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
)

// NewAppLogger monta o logger do processo: texto em stderr, mais um tee para os
// structured logs do Sentry a partir de Warn.
//
// ARMADILHA: o handler base é construído aqui de propósito. Embrulhar
// slog.Default().Handler() e depois chamar slog.SetDefault trava o processo na
// primeira linha de log — o handler padrão escreve pelo pacote log, o
// SetDefault manda o pacote log de volta para o handler padrão, e o mutex do
// log fecha contra si mesmo. Sem panic e sem mensagem.
func NewAppLogger(dsn string) *slog.Logger {
	base := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(NewSentrySlogHandler(base, slog.LevelWarn, dsn))
}

// NewSentrySlogHandler duplica os registros: sempre para base, e de minLevel
// para cima também para o Sentry. Sem DSN devolve base sem alteração.
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
		return e.String(key, fmt.Sprint(v.Any()))
	}
}
