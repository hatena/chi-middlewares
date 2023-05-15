package middleware

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slog"
)

func Test_RequestLogger(t *testing.T) {
	type reqEntry struct {
		Level  string `json:"level"`
		Msg    string `json:"msg"`
		Method string `json:"method"`
		Uri    string `json:"uri"`
		Bytes  int64  `json:"bytes"`
		Status int    `json:"status"`
	}

	type panicEntry struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Panic string `json:"panic"`
	}

	t.Run("HTTP request の要約を log に吐く", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		path := fmt.Sprintf("/%d", rand.Int63())

		r := chi.NewRouter()
		r.Use(Logger(log))
		r.Use(RequestLogger())
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("")) // nolint:errcheck
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("status is not OK: %s", w.Result().Status)
		}

		got := reqEntry{}
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := reqEntry{
			Level:  "INFO",
			Msg:    "request complete",
			Method: "GET",
			Uri:    "http://example.com" + path + " HTTP/1.1",
			Bytes:  0,
			Status: 200,
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("HTTPS request の要約を log に吐く", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		path := fmt.Sprintf("/%d", rand.Int63())

		r := chi.NewRouter()
		r.Use(Logger(log))
		r.Use(RequestLogger())
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("")) // nolint:errcheck
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.TLS = &tls.ConnectionState{}
		r.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("status is not OK: %s", w.Result().Status)
		}

		got := reqEntry{}
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := reqEntry{
			Level:  "INFO",
			Msg:    "request complete",
			Method: "GET",
			Uri:    "https://example.com" + path + " HTTP/1.1",
			Bytes:  0,
			Status: 200,
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("panic した log を吐く", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		message := fmt.Sprintf("%d", rand.Int63())

		r := chi.NewRouter()
		r.Use(Logger(log))
		r.Use(RequestLogger())
		r.Use(chiMiddleware.Recoverer)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			panic(message)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

		if w.Result().StatusCode != http.StatusInternalServerError {
			t.Errorf("status is not InternalServerError: %s", w.Result().Status)
		}

		lines := bytes.Split(buf.Bytes(), []byte("\n"))

		got1 := panicEntry{}
		if err := json.Unmarshal(lines[0], &got1); err != nil {
			t.Fatal(err)
		}
		want1 := panicEntry{
			Level: "ERROR",
			Msg:   "panic",
			Panic: message,
		}
		if diff := cmp.Diff(want1, got1); diff != "" {
			t.Error(diff)
		}

		got2 := reqEntry{}
		if err := json.Unmarshal(lines[1], &got2); err != nil {
			t.Fatal(err)
		}
		want2 := reqEntry{
			Level:  "INFO",
			Msg:    "request complete",
			Method: "GET",
			Uri:    "http://example.com/ HTTP/1.1",
			Bytes:  0,
			Status: 500,
		}
		if diff := cmp.Diff(want2, got2); diff != "" {
			t.Error(diff)
		}
	})
}
