// logger package は、Logger middleware と RequestLogger middleware を提供する
package logger

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type LogKey struct{}

// Logger は、http.Request の Context に slog.Logger を格納する。格納した slog.Logger は GetLogger で取り出せる。middleware.RequestID は Logger の前に置かねばならない
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log2 := log
			if reqID := middleware.GetReqID(ctx); reqID != "" { // RequestID はこの前に適用する事
				log2 = log2.With(slog.String("req_id", reqID))
			}
			ctx = NewContext(ctx, log2)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// NewContext は、slog.Logger を context.Context に格納する。GetLogger で取り出せる
func NewContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, LogKey{}, log)
}

// GetLogger は、Logger で context に格納した slog.Logger を取り出す。slog.Logger を格納していなかったら、slog.Default() を返す
func GetLogger(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(LogKey{}).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
