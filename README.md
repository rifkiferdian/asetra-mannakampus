# Stok Hadiah

Aplikasi web sederhana untuk manajemen stok hadiah dengan autentikasi user berbasis session menggunakan Gin (Golang).

## Tech stack

- Go (Gin Web Framework)
- MySQL (atau database SQL lain yang kompatibel dengan driver `go-sql-driver/mysql`)
- HTML template (`templates/`) dan asset statis (`assets/`)

## Fitur

- Login dan logout user dengan password yang di-hash (bcrypt)
- Registrasi user baru
- Proteksi halaman menggunakan session (middleware auth)
- Halaman dashboard dasar setelah login

## Struktur Proyek

- [`main.go`](main.go:1) – entry point aplikasi, inisialisasi Gin, session, template, statis, dan start server
- [`routes/web.go`](routes/web.go:1) – definisi route utama (auth, dashboard)
- `controllers/` – handler HTTP (login, register, dashboard, render template)
- `middleware/` – middleware autentikasi dan user session
- `templates/` – file HTML template (login, layout, dashboard, dll.)
- `assets/` – file CSS, JS, dan aset frontend lainnya
- [`go.mod`](go.mod:1) – dependensi Go module

## Persyaratan

- Go 1.21+ terinstall
- MySQL server berjalan

## Konfigurasi Environment

Aplikasi menggunakan package [`github.com/joho/godotenv`](go.mod:23) untuk membaca konfigurasi dari file `.env` (jika ada) dan variabel environment OS.

Buat file `.env` di root proyek dengan isi kira‑kira seperti di bawah ini (sesuaikan dengan environment Anda):

```env
APP_PORT=8080

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=password_anda
DB_NAME=stok_hadiah
```

Nilai di atas contoh saja; cek implementasi di package `config` untuk memastikan nama variabel yang digunakan.

## Menjalankan Aplikasi

1. Clone repository ini
2. Masuk ke folder proyek:

   ```bash
   cd gobase-app
   ```

3. Download dependensi:

   ```bash
   go mod tidy
   ```

4. Jalankan aplikasi:

   ```bash
   go run main.go
   ```

5. Buka browser dan akses:

   ```text
   http://localhost:8080
   ```

Atau sesuaikan dengan nilai `APP_PORT` yang Anda gunakan.

## Endpoint Utama

- `GET /` atau `GET /login` – halaman login
- `POST /login` – proses login
- `POST /register` – registrasi user baru
- `GET /logout` – logout user
- `GET /dashboard` – halaman dashboard (butuh login, dilindungi middleware)

Definisi route dapat dilihat di [`routes/web.go`](routes/web.go:10).

## Session & Autentikasi

Aplikasi menggunakan session berbasis cookie dari package [`github.com/gin-contrib/sessions`](go.mod:11) dengan store cookie default:

- Session name: `mysession`
- Key utama (secret): diset langsung di [`main.go`](main.go:31) – sebaiknya diubah dan disimpan di environment untuk production.

Middleware autentikasi dan pengambilan informasi user didefinisikan di package `middleware` dan digunakan di [`routes/web.go`](routes/web.go:19).

## Lisensi

Proyek ini digunakan untuk kebutuhan internal / pembelajaran. Silakan modifikasi sesuai kebutuhan Anda.
