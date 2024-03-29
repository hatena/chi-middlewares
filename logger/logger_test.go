package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		r.Use(middleware.RequestID)
		r.Use(Logger(log))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			reqID = middleware.GetReqID(ctx)
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

func Test_NewContext(t *testing.T) {
	t.Run("slog.Logger を格納する", func(t *testing.T) {
		ctx := context.Background()
		log := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		ctx = NewContext(ctx, log)

		log2, ok := ctx.Value(LogKey{}).(*slog.Logger)
		if !ok {
			t.Fatal("logger is not in the context")
		}

		if log != log2 {
			t.Error("got a different logger")
		}
	})
}

func Test_GetLogger(t *testing.T) {
	t.Run("格納した slog.Logger を取り出せる", func(t *testing.T) {
		log := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))

		ctx := context.Background()
		ctx = context.WithValue(ctx, LogKey{}, log)

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
