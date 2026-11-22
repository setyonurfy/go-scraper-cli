package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		baseURL      string
		wantTitle    string
		wantLinkURLs []string
		wantTypes    []string
	}{
		{
			name: "ekstrak title, normalisasi relatif, filter invalid, dan klasifikasi internal/external",
			html: `
				<html>
				<head><title>  Contoh Halaman  </title></head>
				<body>
					<a href="/about">Tentang</a>
					<a href="/about#tim">Tim (duplikat setelah fragment dihapus)</a>
					<a href="https://example.com/kontak">Kontak</a>
					<a href="https://luar.org/artikel">Artikel Luar</a>
					<a href="#bagian-1">Anchor saja</a>
					<a href="javascript:void(0)">JS</a>
					<a href="mailto:test@example.com">Email</a>
					<a href="">Kosong</a>
				</body>
				</html>
			`,
			baseURL:   "https://example.com/beranda",
			wantTitle: "Contoh Halaman",
			wantLinkURLs: []string{
				"https://example.com/about",
				"https://example.com/kontak",
				"https://luar.org/artikel",
			},
			wantTypes: []string{
				"internal",
				"internal",
				"external",
			},
		},
		{
			name:         "halaman tanpa title dan tanpa link",
			html:         `<html><body><p>Halo dunia</p></body></html>`,
			baseURL:      "https://example.com/",
			wantTitle:    "",
			wantLinkURLs: []string{},
			wantTypes:    []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.html, tc.baseURL)
			if err != nil {
				t.Fatalf("Parse() mengembalikan error tak terduga: %v", err)
			}

			if got.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tc.wantTitle)
			}

			if len(got.Links) != len(tc.wantLinkURLs) {
				t.Fatalf("len(Links) = %d, want %d (%+v)", len(got.Links), len(tc.wantLinkURLs), got.Links)
			}

			for i, wantURL := range tc.wantLinkURLs {
				if got.Links[i].URL != wantURL {
					t.Errorf("Links[%d].URL = %q, want %q", i, got.Links[i].URL, wantURL)
				}
				if got.Links[i].Type != tc.wantTypes[i] {
					t.Errorf("Links[%d].Type = %q, want %q", i, got.Links[i].Type, tc.wantTypes[i])
				}
			}
		})
	}
}
