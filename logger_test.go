package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func Test_Logger(t *testing.T) {
	t.Run("指定した slog.Logger を取り出せる", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		message := fmt.Sprintf("%d", rand.Int63())

		r := chi.NewRouter()
		r.Use(Logger(log))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			log := GetLogger(r.Context())
			log.Info(message)
			w.Write([]byte("")) // nolint:errcheck
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("status is not OK: %s", w.Result().Status)
		}
		if got := buf.String(); !strings.Contains(got, `"msg":"`+message) {
			t.Errorf("got is %s, dose not contain %s", got, message)
		}
	})

	t.Run("ReqID を log に吐く", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		var reqID string

		r := chi.NewRouter()
		r.Use(chiMiddleware.RequestID)
		r.Use(Logger(log))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			reqID = chiMiddleware.GetReqID(ctx)
			log := GetLogger(ctx)
			log.Info("")
			w.Write([]byte("")) // nolint:errcheck
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("status is not OK: %s", w.Result().Status)
		}
		if got := buf.String(); !strings.Contains(got, `"req_id":"`+reqID) {
			t.Errorf("got is %s, dose not contain %s", got, reqID)
		}
	})
}

func Test_GetLogger(t *testing.T) {
	t.Run("格納した slog.Logger を取り出せる", func(t *testing.T) {
		log := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))

		ctx := context.Background()
		ctx = context.WithValue(ctx, logKey{}, log)

		got := GetLogger(ctx)
		if log != got {
			t.Error("got a different logger")
		}
	})

	t.Run("slog.Logger を格納していなければ、slog.Default() を取得できる", func(t *testing.T) {
		log := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		slog.SetDefault(log)

		ctx := context.Background()

		got := GetLogger(ctx)
		if log != got {
			t.Error("got a different logger")
		}
	})
}
