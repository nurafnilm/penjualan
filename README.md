# Penjualan

## Deskripsi
Backend sederhana untuk transaksi penjualan produk. Full CRUD dengan Gin, GORM (Postgres), dan Swagger docs. Endpoint: /api/v1/transactions.

## Installation
1. Clone
2. `cd backend-penjualan`
3. `go mod tidy`
4. Setup Postgres: Buat DB salesdb
5. Update `.env` di folder backend-penjualan.

__.env Example__
```DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_pass
DB_NAME=salesdb
DB_PORT=5432 (default, sesuaikan)```

## Run
- go run main.go
- Swagger: http://localhost:8080/swagger/index.html

## Endpoints
- GET /api/v1/transactions: List (filter berdasarkan product_id=xxx atau start_date=YYYY-MM-DD)
- POST /api/v1/transactions: Buat baru (body: {"product_id": "xxx", "quantity": 2, "price": 15000000})
- PATCH /api/v1/transactions/{id}: Update partial (body: {"quantity": 3})
- DELETE /api/v1/transactions/{id}: Hapus (berdasarkan id)

## Screenshoot Percobaan

### Tampilan Awal
<img width="1920" height="1080" alt="Screenshot 2025-12-19 140848" src="https://github.com/user-attachments/assets/7bbdfd14-13c6-401c-859e-a5305ea9672a" />

### GET 
Contoh filter berdasarkan produk saja:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133649" src="https://github.com/user-attachments/assets/41a95236-b3e3-4324-a821-51ee8b3e57d9" />
Hasilnya:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133701" src="https://github.com/user-attachments/assets/6851f0d6-9372-454d-b8ec-4c171de5dead" />

### POST
Post data penjualan (nama, total, harga):
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133100" src="https://github.com/user-attachments/assets/1c509299-b950-4e38-80d2-5b0df8b5099b" />
Hasilnya di pgAdmin:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133014" src="https://github.com/user-attachments/assets/7ab2da93-0fe6-438e-b18d-fd5aed88b4de" />

### PATCH
Mencoba mengubah data:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133113" src="https://github.com/user-attachments/assets/a0a75dc5-c92a-4bdb-a4ad-df90314232bb" />
Hasilnya (produk id yang kedua berubah, bisa dibandingkan dengan screenshoot pgAdmin pada post:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133025" src="https://github.com/user-attachments/assets/6b85d55e-b089-4197-a64a-4ff992832d36" />

### DELETE
Delete berdasarkan ID:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133147" src="https://github.com/user-attachments/assets/b339bb0e-f3b1-4c84-8d18-c73cb40153d2" />
Hasil di PGAdmin:
<img width="1920" height="1080" alt="Screenshot 2025-12-19 133202" src="https://github.com/user-attachments/assets/368a9424-ce6e-4fa4-bcaa-350801d2a3bd" />
Ketika disearch tidak muncul:
<img width="1920" height="1080" alt="image" src="https://github.com/user-attachments/assets/29386f05-1779-452b-8936-27c2e56a2319" />


