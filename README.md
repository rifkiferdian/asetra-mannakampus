# Asetra

Asetra adalah aplikasi web internal untuk proses pembelian perusahaan dan pengelolaan asset dengan alur utama:

`PR -> Approval -> PO -> GR -> Invoice -> Payment -> Asset`

Fokus aplikasi ini:

- kontrol approval yang jelas dan bertahap
- budget control per store/divisi/GL account
- audit trail untuk setiap aksi penting
- attachment dokumen pada setiap proses pembelian
- pencatatan asset untuk item CAPEX yang sudah diterima

## Cakupan Modul

Modul bisnis yang menjadi target aplikasi:

- Purchase Request (`purchase_requests`, `purchase_request_items`)
- Approval engine (`approval_rules`, `approval_rule_steps`, `approvals`, `approval_tasks`)
- Purchase Order (`purchase_orders`, `purchase_order_items`)
- Goods Receipt (`goods_receipts`, `goods_receipt_items`)
- Invoice dan 3-way matching (`invoices`, `invoice_items`)
- Payment (`payments`, `payment_status_histories`)
- Budget (`budgets`, `budget_usages`)
- Asset (`assets`, `asset_movements`, `asset_documents`)
- Attachment dan audit (`attachments`, `audit_logs`)

## Status Implementasi Saat Ini

Codebase Go yang aktif saat ini baru mencakup fondasi aplikasi dan beberapa modul admin:

- autentikasi login/logout berbasis session cookie
- dashboard
- manajemen user
- manajemen role dan permission
- manajemen store
- template layout dan asset frontend

Skema database contoh untuk domain PO dan asset sudah tersedia di [gobase_app.sql](gobase_app.sql:1), tetapi tidak semua modul bisnis tersebut sudah terhubung ke route, controller, service, dan repository.

## Tech Stack

- Go
- Gin Web Framework
- MySQL
- HTML template server-side
- Tailwind CSS
- Session cookie authentication

## Struktur Proyek

- [main.go](main.go:1) entry point aplikasi
- [routes/web.go](routes/web.go:1) route HTTP utama
- [config/db.go](config/db.go:1) inisialisasi koneksi database
- [controllers](controllers) handler request
- [services](services) business logic per modul
- [repositories](repositories) akses database
- [models](models) model dan DTO
- [middleware](middleware) auth dan permission middleware
- [templates](templates) HTML template
- [assets](assets) CSS, JS, font, dan static assets
- [gobase_app.sql](gobase_app.sql:1) skema dan seed database awal

## Persyaratan

- Go 1.21 atau lebih baru
- MySQL server aktif
- Node.js dan npm untuk build asset CSS

## Konfigurasi Environment

Aplikasi membaca konfigurasi dari file `.env` menggunakan `godotenv`.

Contoh konfigurasi:

```env
APP_NAME=Asetra
APP_PORT=8083
# BASE_URL=http://localhost:8083

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=
DB_NAME=asetra_manna_kampus
```

Variabel database yang dipakai aplikasi saat ini didefinisikan di [config/db.go](config/db.go:1):

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASS`
- `DB_NAME`

## Setup Database

1. Buat database MySQL, misalnya `asetra_manna_kampus`.
2. Import file [gobase_app.sql](gobase_app.sql:1).
3. Sesuaikan `.env` agar mengarah ke database tersebut.

Contoh:

```sql
CREATE DATABASE asetra_manna_kampus;
```

Lalu import:

```bash
mysql -u root -p asetra_manna_kampus < gobase_app.sql
```

## Menjalankan Aplikasi

1. Install dependency Go:

```bash
go mod tidy
```

2. Install dependency frontend:

```bash
npm install
```

3. Build CSS:

```bash
npm run build:css
```

4. Jalankan aplikasi:

```bash
go run main.go
```

5. Buka browser ke:

```text
http://localhost:8083
```

Sesuaikan port jika `APP_PORT` berbeda.

## Script Frontend

Script yang tersedia di [package.json](package.json:1):

- `npm run build:css` untuk build Tailwind CSS sekali jalan
- `npm run watch:css` untuk mode watch selama development

## Endpoint Dasar

Endpoint yang sudah aktif saat ini:

- `GET /` atau `GET /login`
- `POST /login`
- `POST /register`
- `GET /logout`
- `GET /dashboard`
- `GET /stores`
- `GET /users`
- `GET /role`

Route lengkap dapat dilihat di [routes/web.go](routes/web.go:1).

## Session dan Security

Aplikasi menggunakan session cookie melalui package `github.com/gin-contrib/sessions`.

Catatan penting:

- nama session: `mysession`
- cookie `HttpOnly` aktif
- `Secure` cookie mengikuti `APP_SECURE_COOKIE=true`
- secret session masih hard-coded di [main.go](main.go:1) dan sebaiknya dipindahkan ke environment untuk production

Untuk modul approval, invoice, payment, dan audit trail, aturan bisnis targetnya mengacu ke dokumen [AGENTS.md](AGENTS.md:1).

## Roadmap Implementasi yang Disarankan

Urutan pengembangan yang paling aman:

1. master data: vendor, GL account, division, store approver
2. approval rule dan approval task engine
3. purchase request
4. purchase order
5. goods receipt
6. invoice dan 3-way matching
7. payment
8. asset CAPEX
9. reporting dan dashboard real-time

## Catatan

Beberapa metadata project masih memakai nama lama di file lain, misalnya [package.json](package.json:1). README ini sudah disesuaikan untuk domain aplikasi yang sekarang.
