package middleware

import (
	"context"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

type logKey struct{}

// Logger は、http.Request の Context に slog.Logger を格納する。格納した slog.Logger は GetLogger で取り出せる。chiMiddleware.RequestID は Logger の前に置かねばならない
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log2 := log
			if reqID := chiMiddleware.GetReqID(ctx); reqID != "" { // RequestID はこの前に適用する事
				log2 = log2.With(slog.String("req_id", reqID))
			}
			ctx = context.WithValue(ctx, logKey{}, log2)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// GetLogger は、Logger で context に格納した slog.Logger を取り出す。slog.Logger を格納していなかったら、slog.Default() を返す
func GetLogger(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(logKey{}).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
