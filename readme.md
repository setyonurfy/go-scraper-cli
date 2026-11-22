# go-scraper-cli

[![CI](https://github.com/<USERNAME>/go-scraper-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/<USERNAME>/go-scraper-cli/actions/workflows/ci.yml)

CLI Web Scraper dan Concurrent Crawler modular berbasis Go dengan *worker pool*, *rate limiting*, pembatasan kedalaman, pengujian otomatis, dan GitHub Actions CI.

## 1. Ringkasan Proyek

**go-scraper-cli** adalah aplikasi baris perintah (CLI) berbasis Go yang dirancang secara modular untuk mengambil halaman web, mengekstrak judul serta tautan HTML, menelusuri tautan internal secara rekursif dengan batas aman, dan mengekspor hasilnya ke format JSON, CSV, maupun tampilan teks di terminal.

## 2. Fitur Utama

- **HTTP Fetcher (`internal/fetcher`):** Mengambil konten HTML menggunakan `net/http` dengan konfigurasi *timeout* per *request* serta header *User-Agent* kustom.
- **HTML Parser (`internal/parser`):** Mengekstrak judul halaman (`<title>`) dan tautan (`<a href>`), menormalkan URL relatif menjadi absolut, menyaring skema non-HTTP (`#`, `mailto:`, `javascript:`), melakukan deduplikasi, serta memisahkan tautan `internal` dan `external`.
- **Concurrent Crawler (`internal/crawler`):** Menjalankan *worker pool* paralel dengan *rate limiting* berbasis `time.Ticker`, pelacakan URL yang sudah dikunjungi (`visited`), pembatalan bersih melalui `context` (`Ctrl+C`), batas kedalaman (`--depth`), dan batas total halaman (`--max-pages`).
- **Multi-Format Exporter (`internal/exporter`):** Menyimpan hasil *scraping* ke berkas JSON atau CSV secara aman, atau menampilkannya langsung sebagai teks di *stdout*.
- **Unit Testing & CI (`.github/workflows/ci.yml`):** Pengujian otomatis memakai `testing`, `httptest.NewServer`, dan `t.TempDir()` dengan deteksi *race condition* (`go test -race`) yang terintegrasi ke GitHub Actions CI.

## 3. Struktur Direktori Proyek

```text
go-scraper-cli/
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── README.md
├── go.mod
├── cmd/
│   └── scraper/
│       └── main.go
└── internal/
    ├── crawler/
    │   ├── crawler.go
    │   └── crawler_test.go
    ├── exporter/
    │   ├── exporter.go
    │   └── exporter_test.go
    ├── fetcher/
    │   ├── fetcher.go
    │   └── fetcher_test.go
    └── parser/
        ├── parser.go
        └── parser_test.go
```

## 4. Daftar Opsi Flag CLI

| Flag | Default | Keterangan |
| --- | --- | --- |
| `--url` | `https://quotes.toscrape.com/` | URL target tunggal untuk proses *scraping* atau titik awal *crawling*. |
| `--urls` | `""` | Daftar beberapa URL dipisahkan tanda koma (menggantikan `--url` jika diisi). |
| `--depth` | `1` | Batas kedalaman penelusuran tautan internal (`1` = halaman awal saja). |
| `--max-pages` | `20` | Batas maksimal total halaman yang diproses (`0` = tanpa batas). |
| `--workers` | `3` | Jumlah *worker* paralel dalam *worker pool*. |
| `--delay` | `300ms` | Jeda minimal antar *request* HTTP untuk mengendalikan beban ke server target. |
| `--timeout` | `10s` | Batas waktu tunggu (*timeout*) untuk setiap permintaan HTTP. |
| `--internal-only` | `false` | Menyaring hasil keluaran agar hanya menyertakan tautan berjenis `internal`. |
| `--format` | `json` | Format keluaran data: `json`, `csv`, atau `text`. |
| `--out` | `hasil.<format>` | Lokasi dan nama berkas tujuan saat menggunakan format `json` atau `csv`. |

## 5. Contoh Perintah Penggunaan

Contoh perintah berikut menggunakan situs latihan publik untuk demonstrasi operasional CLI:

### A. Scrape Satu Halaman ke Format JSON

```bash
go run ./cmd/scraper --url https://quotes.toscrape.com/ --format json --out hasil.json
```

### B. Scrape Beberapa URL Secara Paralel ke Format CSV

```bash
go run ./cmd/scraper \
  --urls "https://quotes.toscrape.com/,https://quotes.toscrape.com/page/2/" \
  --workers 2 \
  --delay 400ms \
  --format csv \
  --out hasil.csv
```

### C. Crawl Rekursif dengan Batas Kedalaman dan Maksimal Halaman

```bash
go run ./cmd/scraper \
  --url https://quotes.toscrape.com/ \
  --depth 2 \
  --max-pages 10 \
  --workers 3 \
  --delay 300ms \
  --internal-only \
  --format json \
  --out crawl.json
```

## 6. Pengujian, Validasi, dan Otomasi CI

Sebelum melakukan *commit* atau *push* ke repositori GitHub, jalankan rangkaian pemeriksaan statis, pengujian konkurensi, dan verifikasi kompilasi berikut:

```bash
go vet ./...
go test -race -v ./...
go build -o /dev/null ./cmd/scraper
```

> **Catatan Praktik Baik:** Selalu gunakan kombinasi `--max-pages` dan `--delay` yang wajar saat menaikkan nilai `--depth` lebih dari 2 agar jumlah permintaan HTTP tetap terkendali dan tidak membebani server target.