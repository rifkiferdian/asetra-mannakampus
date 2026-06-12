# AGENTS.md — Purchase Order (PR→PO→GR→Invoice→Payment) + Asset

Dokumen ini menjelaskan **alur bisnis**, **modul**, **status**, **tabel data**, dan **aturan keamanan** untuk aplikasi Purchase Order dan Asset.
Tujuan: membantu AI/agent memahami aplikasi dengan cepat untuk:
- membaca konteks domain
- memahami flow dari PR sampai pembayaran
- memahami kontrol budget, approval, dan audit trail
- memahami manajemen asset (CAPEX) dari PO/GR

---

## 1) Ringkasan Sistem

Aplikasi ini menangani proses pembelian perusahaan dengan alur:

1. **PR (Purchase Request)** dibuat oleh requester (store/divisi).
2. **Approval PR** otomatis berdasarkan rule (nominal, spend_type, urgent, dsb).
3. **PO (Purchase Order)** dibuat dari PR yang sudah approved.
4. **GR (Goods Receipt)** dibuat saat barang datang (bisa partial).
5. **Invoice** dicatat saat vendor mengirim tagihan.
6. **3-Way Matching** (PO vs GR vs Invoice).
7. **Payment** dibuat & status pembayaran dikelola.
8. Untuk CAPEX: barang yang diterima dapat dibuat sebagai **Asset**.

Fokus utama:
- Approval cepat & transparan
- Budget control
- Audit trail lengkap (siapa melakukan apa & kapan)
- Attachment rapi untuk setiap dokumen

---

## 2) Aktor / Role Sistem

Role (contoh, sesuaikan dengan Spatie roles di DB):
- `super-admin`
- `admin`
- `requester`
- `manager` (Store Manager)
- `ga-manager` (opsional)
- `it-manager` (opsional)
- `finance-manager`
- `gm` (opsional)
- `procurement`
- `warehouse` (penerima barang)

Catatan:
- Sistem menggunakan Spatie: `roles`, `model_has_roles`, `permissions` untuk akses fitur.
- Untuk approval store tanpa area, gunakan mapping approver per store: `store_approvers`.

---

## 3) Modul Utama & Tujuan

### 3.1 PR (Purchase Request)
- Requester membuat PR + item.
- PR adalah estimasi kebutuhan, bisa OPEX/CAPEX.
- PR masuk approval otomatis.

### 3.2 Approval (RULE + STEP + TASK)
- Rule memilih jalur approval.
- Steps menentukan urutan role approver.
- Tasks adalah tugas approve nyata per user.

### 3.3 PO (Purchase Order)
- PO wajib dibuat dari PR yang sudah `APPROVED`.
- PO berisi harga final dari procurement/vendor.

### 3.4 GR (Goods Receipt)
- GR dibuat saat barang diterima.
- Bisa partial (beberapa GR untuk 1 PO).
- GR menjadi dasar update stok/asset.

### 3.5 Invoice + 3-Way Matching
- Invoice dicatat per PO.
- Matching memverifikasi qty & harga sesuai PO dan qty diterima sesuai GR.
- Invoice dapat `DISPUTED` bila ada mismatch.

### 3.6 Payment
- Payment dibuat setelah invoice `MATCHED` / `APPROVED_FOR_PAYMENT`.
- Payment status dicatat lengkap dengan history.

### 3.7 Asset (CAPEX)
- Item CAPEX yang diterima via GR dapat dibuat menjadi asset:
  - assign kode asset
  - lokasi (store)
  - PIC
  - nilai perolehan
  - tanggal perolehan
  - warranty (opsional)
  - depresiasi (opsional)

### 3.8 Budget Control
- Budget diset per store/divisi/GL/periode.
- Pemakaian budget dicatat (rekomendasi: saat PO `APPROVED`).

### 3.9 Attachments & Audit
- Attachment untuk bukti (foto, PDF, quotation, invoice scan, DO).
- Audit log untuk semua aksi penting.

---

## 4) Status Enum (Standar)

### 4.1 PR status
- `DRAFT`
- `SUBMITTED`
- `IN_APPROVAL`
- `REJECTED`
- `APPROVED`
- `CONVERTED_TO_PO`
- `CLOSED`

### 4.2 Approval status
- approvals.status: `PENDING`, `APPROVED`, `REJECTED`, `CANCELLED`
- approval_tasks.status: `WAITING`, `APPROVED`, `REJECTED`, `SKIPPED`

### 4.3 PO status
- `DRAFT`
- `SUBMITTED`
- `IN_APPROVAL`
- `REJECTED`
- `APPROVED`
- `RECEIVING`
- `CLOSED`

### 4.4 GR status
- `DRAFT`
- `POSTED`

### 4.5 Invoice status
- `RECEIVED`
- `MATCHED`
- `DISPUTED`
- `APPROVED_FOR_PAYMENT`
- `PAID`

### 4.6 Payment status
- `PENDING`
- `PAID`
- `FAILED`

### 4.7 GL spend_type
- `OPEX`
- `CAPEX`

---

## 5) Aturan Bisnis Kunci

### 5.1 PR → PO wajib
- PO tidak boleh dibuat bila PR belum `APPROVED`.
- Disarankan enforce di aplikasi + trigger (opsional): block insert PO jika PR not approved.

### 5.2 PR amount vs PO amount
- PR umumnya estimasi.
- PO umumnya nilai final.
- Disarankan aturan tolerance:
  - jika PO > PR dan selisih > X% => approval ulang / escalate.

### 5.3 Budget check
- Budget check saat PR submit (cek sisa) dan final di PO approved.
- Pemakaian budget dicatat saat PO `APPROVED` (simple & umum).

### 5.4 GR partial
- 1 PO dapat memiliki banyak GR.
- PO dapat ditutup jika semua item qty_received == qty_ordered.

### 5.5 3-way matching
- Invoice boleh lanjut bayar hanya jika:
  - qty invoice <= qty received (GR)
  - harga invoice <= harga PO (atau sesuai policy)
- Jika mismatch → `DISPUTED`.

---

## 6) Data Model (Ringkas)

> Nama tabel mengikuti desain yang dipakai dalam diskusi.

### 6.1 Master
- `users` (existing)
- `roles`, `model_has_roles` (Spatie existing)
- `stores` (existing)
- `user_stores` (mapping akses store)
- `store_approvers` (mapping approver per store-role)
- `divisions` (opsional, untuk grouping)
- `gl_accounts` (dengan spend_type OPEX/CAPEX)
- `vendors`

### 6.2 PR
- `purchase_requests`
- `purchase_request_items`

### 6.3 Approval
- `approval_rules`
- `approval_rule_steps`
- `approvals`
- `approval_tasks`

### 6.4 PO & GR
- `purchase_orders`
- `purchase_order_items`
- `goods_receipts`
- `goods_receipt_items`

### 6.5 Invoice & Payment
- `invoices`
- `invoice_items`
- `payments`
- `invoice_status_histories`
- `payment_status_histories`

### 6.6 Budget
- `budgets`
- `budget_usages`

### 6.7 Transparansi
- `attachments` (ref_type + ref_id)
- `audit_logs` (ref_type + ref_id)

### 6.8 Asset (CAPEX)
- `assets` (asset master)
- `asset_movements` (mutasi/lokasi)
- `asset_documents` (lampiran khusus asset, opsional; atau gunakan attachments ref_type='ASSET')

---

## 7) Alur Detail (Step-by-step)

### 7.1 Buat PR
1. Requester isi header PR + items.
2. Sistem hitung total_amount dari items.
3. Submit: PR status -> `SUBMITTED` lalu generate approval:
   - approvals + approval_tasks dibuat dari rule.
4. Audit log: PR SUBMIT.

### 7.2 Approval PR
1. Approver membuka inbox tasks yang `WAITING`.
2. Approve/Reject:
   - Validasi backend: task milik user (assigned_user_id == current_user_id).
   - Validasi step sebelumnya selesai.
3. Tulis history:
   - update approval_tasks
   - insert audit_logs (APPROVE/REJECT)
4. Jika semua step selesai:
   - approvals.status = APPROVED
   - PR.status = APPROVED

### 7.3 Buat PO dari PR (Procurement)
1. PO dibuat dari PR approved:
   - copy items dari PR items ke PO items
   - set vendor
2. PO status -> `APPROVED` (atau melalui approval PO jika dibutuhkan)
3. Budget usage dicatat saat PO approved (recommended):
   - insert budget_usages (ref_type='PO', ref_id=po_id)
4. PR status -> `CONVERTED_TO_PO`

### 7.4 GR (Receiving)
1. Saat barang datang, warehouse/store membuat GR:
   - input qty_received per po_item_id
2. Jika menerima sebagian -> buat GR berikutnya.
3. Bila semua item terpenuhi -> PO status -> `CLOSED`.
4. Untuk CAPEX item -> create asset setelah GR posted.

### 7.5 Invoice + Matching
1. Finance input invoice header + invoice_items (mengacu ke po_item_id).
2. Matching:
   - PO qty vs GR qty vs Invoice qty
   - PO price vs Invoice price
3. Jika OK -> invoice status `MATCHED`.
4. Jika mismatch -> `DISPUTED`.

### 7.6 Payment
1. Jika invoice `MATCHED`, finance set `APPROVED_FOR_PAYMENT`.
2. Buat payment record `PENDING`.
3. Setelah transfer -> payment `PAID`, invoice `PAID`.
4. Semua perubahan status wajib masuk history.

---

## 8) Security & Integrity Checklist (Wajib)

### 8.1 Approval anti manipulasi
- Client tidak boleh mengirim status langsung.
- Endpoint khusus: `/tasks/{id}/approve` & `/tasks/{id}/reject`.
- Backend wajib cek:
  - task assigned ke user login
  - task status WAITING
  - step gating (prev step approved)
  - transaksi + locking (SELECT FOR UPDATE)
- Log semua aksi ke `audit_logs`.

### 8.2 Invoice/Payment status change
- Update status hanya lewat service (transaction):
  - read old_status
  - update new_status
  - insert history (invoice_status_histories / payment_status_histories)
  - insert audit_logs
- Catat ip_address + user_agent (opsional tapi recommended).

### 8.3 DB access
- DB tidak diekspos publik.
- gunakan credential & privilege minimum untuk app user.

---

## 9) Attachment Rules
- File disimpan di storage (filesystem/S3), DB hanya simpan metadata:
  - file_path, file_name, mime_type, file_size
- Attachment dikelompokkan dengan `ref_type` + `ref_id`:
  - PR: lampiran permintaan, foto, TOR
  - PO: quotation, kontrak
  - GR: DO, foto penerimaan
  - INVOICE: scan invoice, faktur pajak

---

## 10) Naming & Conventions
- Nomor dokumen:
  - PR: `PR-{STORECODE}-{YYYY}-{SEQ}`
  - PO: `PO-{STORECODE}-{YYYY}-{SEQ}`
  - GR: `GR-{STORECODE}-{YYYY}-{SEQ}`
  - INV: `INV-{VENDORCODE}-{SEQ}` (opsional)
- Semua perubahan status harus:
  1) update tabel utama
  2) insert history status (jika ada)
  3) insert audit log

---

## 11) Minimal Seed Data untuk Demo
- 1 store MK4
- user requester
- user store manager MK4 (mapping store_approvers)
- user finance-manager
- vendor
- GL accounts: ATK (OPEX), Maintenance (OPEX), Asset IT (CAPEX)

---

## 12) TODO (Jika ingin dikembangkan)
- SLA & escalation approval
- Delegasi approver
- Notifikasi (email/WA)
- Toleransi PR vs PO (variance)
- Asset depreciation schedule
- Integration ke accounting/ERP

---