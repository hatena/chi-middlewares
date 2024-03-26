package logger

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger は、HTTP request が完了した時にその要約を log に吐く。Logger は RequestLogger の前に置かねばならない。middleware.RealIP は RequestLogger の前に置かねばならない。middleware.Recoverer は RequestLogger の後に置かねばならない
var RequestLogger func(http.Handler) http.Handler = middleware.RequestLogger(&requestLogger{})

type requestLogger struct{}

func (l *requestLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	ctx := r.Context()

	log := GetLogger(ctx)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return &requestLogEntry{
		// Panic() で以下を印字しない為に slog.Logger.With() に載せずに渡す
		attrs: []slog.Attr{
			slog.String("method", r.Method),
			slog.String("remote_addr", r.RemoteAddr), // RealIP はこの前に適用する事
			slog.String("uri", fmt.Sprintf(`%s://%s%s %s`, scheme, r.Host, r.RequestURI, r.Proto)),
		},
		ctx: ctx,
		log: log,
	}
}

type requestLogEntry struct {
	attrs []slog.Attr
	ctx   context.Context
	log   *slog.Logger
}

func (e *requestLogEntry) Write(
	status, bytes int,
	header http.Header,
	elapsed time.Duration,
	extra interface{},
) {
	e.log.LogAttrs(
		e.ctx,
		slog.LevelInfo,
		"request complete",
		append(
			e.attrs,
			slog.Float64("elapsed_ms", float64(elapsed.Nanoseconds())/1000000.0),
			slog.Int("bytes", bytes),
			slog.Int("status", status),
		)...,
	)
}

// middleware.Recoverer が呼ぶ
func (e *requestLogEntry) Panic(v interface{}, stack []byte) {
	e.log.ErrorContext(
		e.ctx,
		"panic",
		slog.String("panic", fmt.Sprintf("%+v", v)),
		slog.String("stack", string(stack)),
	)
}
