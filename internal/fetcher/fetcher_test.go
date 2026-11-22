package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetcher_FetchHTML(t *testing.T) {
	t.Run("sukses 200 OK dan mengirim User-Agent", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ua := r.Header.Get("User-Agent"); !strings.Contains(ua, "GoScraperCLI") {
				t.Errorf("User-Agent = %q, diharapkan mengandung GoScraperCLI", ua)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><head><title>OK</title></head></html>"))
		}))
		defer srv.Close()

		f := New(2*time.Second, false)
		body, err := f.FetchHTML(srv.URL)
		if err != nil {
			t.Fatalf("FetchHTML() error = %v", err)
		}
		if !strings.Contains(body, "<title>OK</title>") {
			t.Errorf("body tidak sesuai: %q", body)
		}
	})

	t.Run("gagal pada status HTTP 404", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		f := New(2*time.Second, false)
		if _, err := f.FetchHTML(srv.URL); err == nil {
			t.Fatal("diharapkan error untuk status 404, tetapi nil")
		}
	})

	t.Run("gagal karena timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(150 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		f := New(30*time.Millisecond, false)
		if _, err := f.FetchHTML(srv.URL); err == nil {
			t.Fatal("diharapkan error timeout, tetapi nil")
		}
	})
}
