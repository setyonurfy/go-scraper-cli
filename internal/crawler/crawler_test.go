package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func newTestSite(t *testing.T) (*httptest.Server, *int, *sync.Mutex) {
	t.Helper()

	var hitCount int
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hitCount++
		mu.Unlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		switch r.URL.Path {
		case "/":
			_, _ = fmt.Fprint(w, `
				<html><head><title>Home</title></head><body>
					<a href="/p1">Page 1</a>
					<a href="/p2">Page 2</a>
					<a href="https://external.example.com/out">External</a>
				</body></html>
			`)
		case "/p1":
			_, _ = fmt.Fprint(w, `
				<html><head><title>Page 1</title></head><body>
					<a href="/">Back Home</a>
					<a href="/d3a">Deep 3A</a>
				</body></html>
			`)
		case "/p2":
			_, _ = fmt.Fprint(w, `
				<html><head><title>Page 2</title></head><body>
					<a href="/d3b">Deep 3B</a>
				</body></html>
			`)
		case "/d3a":
			_, _ = fmt.Fprint(w, `<html><head><title>Deep 3A</title></head><body><a href="/d4">Deep 4</a></body></html>`)
		case "/d3b":
			_, _ = fmt.Fprint(w, `<html><head><title>Deep 3B</title></head><body></body></html>`)
		default:
			_, _ = fmt.Fprintf(w, `<html><head><title>%s</title></head><body></body></html>`, r.URL.Path)
		}
	}))

	t.Cleanup(srv.Close)
	return srv, &hitCount, &mu
}

func countSuccess(outcomes []Outcome) int {
	ok := 0
	for _, o := range outcomes {
		if o.Err == nil && o.Result != nil {
			ok++
		}
	}
	return ok
}

func TestRun_MaxDepth(t *testing.T) {
	srv, hitCount, mu := newTestSite(t)

	tests := []struct {
		name     string
		maxDepth int
		wantPage int
	}{
		{
			name:     "depth 1 hanya mengambil halaman awal",
			maxDepth: 1,
			wantPage: 1, // /
		},
		{
			name:     "depth 2 mengikuti link internal tingkat pertama",
			maxDepth: 2,
			wantPage: 3, // /, /p1, /p2 (link eksternal diabaikan)
		},
		{
			name:     "depth 3 menelusuri hingga tingkat kedua dan tidak mengulang URL visited",
			maxDepth: 3,
			wantPage: 5, // /, /p1, /p2, /d3a, /d3b (/ sudah visited)
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			*hitCount = 0
			mu.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			outcomes := Run(ctx, []string{srv.URL + "/"}, Config{
				Workers:  2,
				Delay:    0,
				Timeout:  2 * time.Second,
				MaxDepth: tc.maxDepth,
				MaxPages: 0, // unlimited untuk fokus menguji MaxDepth
			})

			gotSuccess := countSuccess(outcomes)
			if gotSuccess != tc.wantPage {
				t.Errorf("jumlah halaman sukses = %d, want %d (total outcomes = %d)", gotSuccess, tc.wantPage, len(outcomes))
			}

			mu.Lock()
			hits := *hitCount
			mu.Unlock()
			if hits != tc.wantPage {
				t.Errorf("jumlah HTTP request ke server = %d, want %d", hits, tc.wantPage)
			}
		})
	}
}

func TestRun_MaxPages(t *testing.T) {
	srv, hitCount, mu := newTestSite(t)

	tests := []struct {
		name     string
		maxDepth int
		maxPages int
		wantPage int
	}{
		{
			name:     "depth 3 tapi max-pages 2 membatasi total halaman menjadi 2",
			maxDepth: 3,
			maxPages: 2,
			wantPage: 2,
		},
		{
			name:     "depth 3 tapi max-pages 3 membatasi total halaman menjadi 3",
			maxDepth: 3,
			maxPages: 3,
			wantPage: 3,
		},
		{
			name:     "max-pages 1 menghentikan penelusuran link lanjutan",
			maxDepth: 3,
			maxPages: 1,
			wantPage: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			*hitCount = 0
			mu.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			outcomes := Run(ctx, []string{srv.URL + "/"}, Config{
				Workers:  2,
				Delay:    0,
				Timeout:  2 * time.Second,
				MaxDepth: tc.maxDepth,
				MaxPages: tc.maxPages,
			})

			gotSuccess := countSuccess(outcomes)
			if gotSuccess != tc.wantPage {
				t.Errorf("jumlah halaman sukses = %d, want %d", gotSuccess, tc.wantPage)
			}

			mu.Lock()
			hits := *hitCount
			mu.Unlock()
			if hits != tc.wantPage {
				t.Errorf("jumlah HTTP request ke server = %d, want %d", hits, tc.wantPage)
			}
		})
	}
}
