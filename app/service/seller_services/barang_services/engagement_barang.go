package seller_service

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services/response_barang_service"

)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Masukan Barang Seller
// Berfungsi untuk melayani seller yang hendak memasukan barang nya ke sistem burung
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanBarangInduk(db *gorm.DB, data PayloadMasukanBarangInduk) *response.ResponseForm {
	services := "MasukanBarang"

	// Validasi kredensial seller
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseMasukanBarangInduk{
				Message: "Gagal Memasukkan Barang karena kredensial seller tidak valid",
			},
		}
	}

	// Set SellerID ke barang induk
	data.BarangInduk.SellerID = data.IdentitasSeller.IdSeller

	// Validasi kelengkapan data barang
	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		log.Printf("[WARN] Data barang tidak lengkap untuk seller ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal: Data barang harus dilengkapi",
		}
	}

	// Cek apakah nama barang sudah terdaftar untuk seller ini
	var namaBarang string
	errNama := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			SellerID:   data.IdentitasSeller.IdSeller,
			NamaBarang: data.BarangInduk.NamaBarang,
		}).
		Select("nama_barang").
		Take(&namaBarang).Error

	if errNama == nil {
		log.Printf("[WARN] Seller ID %d sudah memiliki barang dengan nama '%s'", data.IdentitasSeller.IdSeller, data.BarangInduk.NamaBarang)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  "Gagal: Nama barang sudah terdaftar untuk seller ini",
		}
	}

	// Jalankan proses insert dalam goroutine (asynchronous)
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			if data.BarangInduk.OriginalKategori == "" {
				return fmt.Errorf("OriginalKategori kosong, rollback transaksi")
			}

			// Insert barang induk
			if err := tx.Create(&data.BarangInduk).Error; err != nil {
				return err
			}

			idInduk := data.BarangInduk.ID

			// Insert kategori barang beserta variannya
			for _, kategori := range data.KategoriBarang {
				kategori.IdBarangInduk = int32(idInduk)

				if err := tx.Create(&kategori).Error; err != nil {
					return err
				}

				idKategori := kategori.ID

				// Siapkan varian batch berdasarkan stok
				var varianBatch []models.VarianBarang
				for i := 0; i < int(kategori.Stok); i++ {
					varianBatch = append(varianBatch, models.VarianBarang{
						IdBarangInduk: int32(idInduk),
						IdKategori:    idKategori,
						Status:        "Ready",
						Sku:           kategori.Sku,
					})
				}

				// Insert varian batch jika ada
				if len(varianBatch) > 0 {
					if err := tx.Create(&varianBatch).Error; err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] MasukanBarang gagal insert data untuk seller ID %d: %v", data.IdentitasSeller.IdSeller, err)
			return
		}

		log.Printf("[INFO] Barang berhasil dimasukkan untuk seller ID %d", data.IdentitasSeller.IdSeller)
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseMasukanBarangInduk{
			Message: "Barang berhasil ditambahkan.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Edit Barang Induk
// Berfungsi untuk seller dalam melakukan edit atau pembaruan informasi seputar barang induknya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditBarangInduk(db *gorm.DB, data PayloadEditBarangInduk) *response.ResponseForm {
	services := "EditBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Gagal: Kredensial Seller Tidak Valid",
			},
		}
	}

	data.BarangInduk.SellerID = data.IdentitasSeller.IdSeller

	var count int64 = 0
	err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID: data.BarangInduk.ID,
		}).
		Count(&count).Error

	if err != nil {
		log.Printf("[ERROR] Terjadi kesalahan server saat validasi barang: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Terjadi Kesalahan Server Saat Validasi Barang",
			},
		}
	}

	if count != 1 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditBarangInduk{
				Message: "Gagal: Barang Tidak Ada atau Tidak Ditemukan",
			},
		}
	}

	go func() {
		data.BarangInduk.TanggalRilis = ""
		data.BarangInduk.Viewed = 0
		data.BarangInduk.Likes = 0
		data.BarangInduk.TotalKomentar = 0
		data.BarangInduk.HargaKategoris = 0

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.BarangInduk{}).
				Where(&models.BarangInduk{
					ID: data.BarangInduk.ID,
				}).
				Updates(&data.BarangInduk).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditBarang gagal update BarangInduk ID %d: %v", data.BarangInduk.ID, err)
			return
		}

		log.Printf("[INFO] Barang berhasil diubah untuk seller ID %d", data.IdentitasSeller.IdSeller)
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarangInduk{
			Message: "Barang Berhasil Diubah.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Barang Induk
// Berfungsi untuk seller dalam menghapus barang induknya, akan otomatis menghapus kategori barang dan varian barang
// nya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusBarangInduk(db *gorm.DB, data PayloadHapusBarangInduk) *response.ResponseForm {
	services := "HapusBarang"

	// Validasi kredensial seller
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarangInduk{
				Message: "Gagal: Kredensial Seller Tidak Valid",
			},
		}
	}

	// Cek apakah masih ada varian dalam transaksi (status Dipesan/Diproses)
	var jumlahDalamTransaksi int64
	if errStock := db.Model(&models.VarianBarang{}).
		Where("id_barang_induk = ? AND status IN ?", data.BarangInduk.ID, []string{"Dipesan", "Diproses"}).
		Count(&jumlahDalamTransaksi).Error; errStock != nil {
		log.Printf("[ERROR] Gagal cek stok dalam transaksi: %v", errStock)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	// Jika masih ada barang dalam transaksi, hentikan proses
	if jumlahDalamTransaksi != 0 {
		log.Printf("[WARN] Masih ada %v barang dalam transaksi untuk barang ID %d", jumlahDalamTransaksi, data.BarangInduk.ID)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarangInduk{
				Message: fmt.Sprintf("Masih ada sejumlah: %v barang dalam transaksi", jumlahDalamTransaksi),
			},
		}
	}

	// Jalankan proses penghapusan dalam goroutine (asynchronous)
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			// Hapus permanen varian barang
			if err := tx.Unscoped().
				Where("id_barang_induk = ?", data.BarangInduk.ID).
				Delete(&models.VarianBarang{}).Error; err != nil {
				return err
			}

			// Hapus soft delete kategori barang
			if err := tx.
				Where("id_barang_induk = ?", data.BarangInduk.ID).
				Delete(&models.KategoriBarang{}).Error; err != nil {
				return err
			}

			// Hapus soft delete barang induk
			if err := tx.
				Where("id = ?", data.BarangInduk.ID).
				Delete(&models.BarangInduk{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] Rollback HapusBarang untuk barang '%s' (ID %d): %v",
				data.BarangInduk.NamaBarang, data.BarangInduk.ID, err)
			return
		}

		log.Printf("[INFO] Barang berhasil dihapus (soft delete) untuk seller ID %d (Barang ID %d)",
			data.IdentitasSeller.IdSeller, data.BarangInduk.ID)
	}()

	// Kembalikan respons sukses
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusBarangInduk{
			Message: "Barang berhasil dihapus.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Tambah Kategori Barang Induk
// Berfungsi untuk seller menambahkan kategori barang pada barang induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahKategoriBarang(db *gorm.DB, data PayloadTambahKategori) *response.ResponseForm {
	services := "TambahKategoriBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_barang_service.ResponseTambahKategori{
				Message: "Gagal, Kredensial seller tidak valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk tambah kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseTambahKategori{
				Message: "Gagal Baang Induk Tidak Ditemukan",
			},
		}
	}

	go func(data PayloadTambahKategori) {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				data.KategoriBarang[i].IdBarangInduk = data.IdBarangInduk

				// //////////////////////////////////////////////////////
				// pencegahan nama kategori yang sama persis
				// //////////////////////////////////////////////////////

				var existingKategori models.KategoriBarang
				err := tx.Where(&models.KategoriBarang{
					Nama: data.KategoriBarang[i].Nama,
					ID:   int64(data.IdBarangInduk),
				}).
					First(&existingKategori).Error

				if err == nil {
					log.Printf("[WARN] Kategori '%s' sudah ada pada barang induk ID %d", data.KategoriBarang[i].Nama, data.IdBarangInduk)
					continue
				} else if err != gorm.ErrRecordNotFound {
					log.Printf("[ERROR] Gagal cek kategori '%s' pada barang induk ID %d: %v", data.KategoriBarang[i].Nama, data.IdBarangInduk, err)
					continue
				}

				if err := tx.Create(&data.KategoriBarang[i]).Error; err != nil {
					log.Printf("[ERROR] Gagal menambah kategori '%s' pada barang induk ID %d: %v", data.KategoriBarang[i].Nama, data.IdBarangInduk, err)
					return err
				}

				var kategoriBaru models.KategoriBarang
				if err := tx.Where(&models.KategoriBarang{
					Nama: data.KategoriBarang[i].Nama,
					ID:   int64(data.IdBarangInduk),
				}).
					First(&kategoriBaru).Error; err != nil {
					log.Printf("[ERROR] Gagal mengambil kategori baru '%s' pada barang induk ID %d: %v", data.KategoriBarang[i].Nama, data.IdBarangInduk, err)
					return err
				}

				var varianBatch []models.VarianBarang
				for j := 0; j < int(data.KategoriBarang[i].Stok); j++ {
					varianBatch = append(varianBatch, models.VarianBarang{
						IdBarangInduk: int32(data.IdBarangInduk),
						IdKategori:    kategoriBaru.ID,
						Status:        "Ready",
						Sku:           data.KategoriBarang[i].Sku,
					})
				}

				if len(varianBatch) > 0 {
					if err := tx.Create(&varianBatch).Error; err != nil {
						log.Printf("[ERROR] Gagal membuat varian untuk kategori '%s' pada barang induk ID %d: %v", data.KategoriBarang[i].Nama, data.IdBarangInduk, err)
						return err
					}
				}
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] TambahKategoriBarang gagal proses kategori & varian untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}(data)

	log.Printf("[INFO] Kategori barang berhasil ditambahkan pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseTambahKategori{
			Message: "Kategori barang berhasil ditambahkan.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Edit Kategori Barang
// Berfungsi untuk mengedit data informasi tentang kategori barang induk yang dituju
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditKategoriBarang(db *gorm.DB, data PayloadEditKategori) *response.ResponseForm {
	services := "EditKategoriBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Baang Induk Tidak Ditemukan",
			},
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				data.KategoriBarang[i].Stok = 0
				if err := tx.Model(&models.KategoriBarang{}).
					Where(&models.KategoriBarang{
						ID: data.KategoriBarang[i].ID,
					}).
					Updates(&data.KategoriBarang[i]).Error; err != nil {
					log.Printf("[ERROR] Gagal update kategori ID %d: %v", data.KategoriBarang[i].ID, err)
					return err
				}

				if err := tx.Model(&models.VarianBarang{}).
					Where(&models.VarianBarang{
						IdBarangInduk: data.IdBarangInduk,
						IdKategori:    data.KategoriBarang[i].ID,
					}).
					Update("sku", data.KategoriBarang[i].Sku).Error; err != nil {
					log.Printf("[ERROR] Gagal update SKU varian untuk kategori ID %d: %v", data.KategoriBarang[i].ID, err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditKategoriBarang gagal update kategori untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil diubah pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditKategori{
			Message: "Kategori barang berhasil diubah.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Kategori Barang Induk
// Berfungsi untuk menghapus kategori barang induk yang ada
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusKategoriBarang(db *gorm.DB, data PayloadHapusKategori) *response.ResponseForm {
	services := "HapusKategoriBarang"
	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk hapus kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Baang Induk Tidak Ditemukan",
			},
		}
	}

	for _, kat := range data.KategoriBarang {
		var barang_dalam_transaksi int64
		if err_stock := db.Model(models.VarianBarang{}).
			Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: kat.ID, Status: "Dipesan"}).
			Or(models.VarianBarang{Status: "Diproses"}).
			Count(&barang_dalam_transaksi).Error; err_stock != nil {
			log.Printf("[ERROR] Gagal cek stok dalam transaksi untuk kategori ID %d: %v", kat.ID, err_stock)
			continue
		}

		if barang_dalam_transaksi != 0 {
			log.Printf("[WARN] Tidak bisa hapus kategori %s, masih ada %v stok dalam transaksi", kat.Nama, barang_dalam_transaksi)
			return &response.ResponseForm{
				Status:   http.StatusConflict,
				Services: services,
				Payload: response_barang_service.ResponseHapusKategori{
					Message: fmt.Sprintf("Tidak Bisa Menghapus Kategori %s, Masih Ada %v Stok Dalam Transaksi", kat.Nama, barang_dalam_transaksi),
				},
			}
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				var existingKategori models.KategoriBarang
				err := tx.Where(&models.KategoriBarang{
					ID:            data.KategoriBarang[i].ID,
					IdBarangInduk: data.IdBarangInduk,
				}).First(&existingKategori).Error

				if err != nil {
					if err == gorm.ErrRecordNotFound {
						log.Printf("[WARN] Kategori ID %d tidak ditemukan untuk barang induk ID %d", data.KategoriBarang[i].ID, data.IdBarangInduk)
						continue
					} else {
						log.Printf("[ERROR] Error cek kategori ID %d: %v", data.KategoriBarang[i].ID, err)
						continue
					}
				}

				if err := tx.Unscoped().
					Where("id_kategori = ? AND id_barang_induk = ?", existingKategori.ID, data.IdBarangInduk).
					Delete(&models.VarianBarang{}).Error; err != nil {
					log.Printf("[ERROR] Gagal hapus varian untuk kategori ID %d: %v", existingKategori.ID, err)
					return err
				}

				if err := tx.Delete(&existingKategori).Error; err != nil {
					log.Printf("[ERROR] Gagal hapus kategori ID %d: %v", existingKategori.ID, err)
					return err
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] HapusKategoriBarang gagal proses hapus kategori & varian untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil dihapus pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusKategori{
			Message: "Kategori barang berhasil dihapus.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// STOK BARANG
// ////////////////////////////////////////////////////////////////////////////////

func EditStokBarang(db *gorm.DB, data PayloadEditStokBarang) *response.ResponseForm {
	services := "EditStokBarang"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var idBarangIndukGet int64 = 0

	if err := db.Model(&models.BarangInduk{}).
		Where(&models.BarangInduk{
			ID:       data.IdBarangInduk,
			SellerID: data.IdentitasSeller.IdSeller,
		}).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk hapus kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	if idBarangIndukGet == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditKategori{
				Message: "Gagal Baang Induk Tidak Ditemukan",
			},
		}
	}

	for _, b := range data.Barang {
		if err := db.
			Where(&models.KategoriBarang{IdBarangInduk: data.IdBarangInduk, ID: b.IdKategoriBarang}).
			Take(&models.KategoriBarang{}).Error; err != nil {
			log.Printf("[WARN] Kategori tidak ditemukan: IdKategori=%d, NamaKategori=%s, Error=%v", b.IdKategoriBarang, b.NamaKategoriBarang, err)
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  fmt.Sprintf("Kategori %s tidak ditemukan", b.NamaKategoriBarang),
			}
		}
		log.Printf("[INFO] Kategori ditemukan: IdKategori=%d, NamaKategori=%s", b.IdKategoriBarang, b.NamaKategoriBarang)
	}

	go func() {
		for _, b := range data.Barang {

			var jumlah int64
			if err := db.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: b.IdKategoriBarang, Status: "Ready"}).
				Count(&jumlah).Error; err != nil {
				log.Printf("[ERROR] Gagal menghitung stok: IdKategori=%d, Error=%v", b.IdKategoriBarang, err)
				continue
			}

			if jumlah == int64(b.JumlahStok) {
				continue
			}

			if jumlah > int64(b.JumlahStok) {
				hapus := jumlah - int64(b.JumlahStok)
				log.Printf("[INFO] Perlu hapus varian: IdKategori=%d, JumlahHapus=%d", b.IdKategoriBarang, hapus)
				for j := int64(0); j < hapus; j++ {
					if err := db.
						Where(&models.VarianBarang{
							IdBarangInduk: data.IdBarangInduk,
							IdKategori:    b.IdKategoriBarang,
							Status:        "Ready",
						}).
						Limit(1).
						Delete(&models.VarianBarang{}).Error; err != nil {
						log.Printf("[ERROR] Gagal menghapus varian: IdKategori=%d, Iterasi=%d, Error=%v", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[INFO] Berhasil menghapus varian: IdKategori=%d, Iterasi=%d", b.IdKategoriBarang, j)
					}
				}
			}

			if int64(b.JumlahStok) > jumlah {
				buat := int64(b.JumlahStok) - jumlah
				for j := int64(0); j < buat; j++ {
					newVarian := models.VarianBarang{
						IdBarangInduk: data.IdBarangInduk,
						IdKategori:    b.IdKategoriBarang,
						Status:        "Ready",
						Sku:           b.SkuKategoriBarang,
					}
					if err := db.Create(&newVarian).Error; err != nil {
						log.Printf("[ERROR] Gagal membuat varian: IdKategori=%d, Iterasi=%d, Error=%v", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[INFO] Berhasil membuat varian: IdKategori=%d, Iterasi=%d", b.IdKategoriBarang, j)
					}
				}
			}
		}
	}()

	log.Printf("[INFO] Proses update stok sedang berjalan untuk barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditStokBarang{
			Message: "Proses update stok sedang berjalan",
		},
	}
}

func DownStokBarangInduk(db *gorm.DB, data PayloadDownBarangInduk) *response.ResponseForm {
	services := "DownStokBarangInduk"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	var count_barang int64
	if err_db := db.Model(&models.BarangInduk{}).Where(&models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdentitasSeller.IdSeller}).Count(&count_barang).Limit(1).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdentitasSeller.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal, barang tidak ditemukan.",
			},
		}
	} else {
		go func() {
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).
					Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: "Ready"}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: "Terjual"}).
					Update("status", "Down").Error; err_updates != nil {
					log.Printf("[ERROR] Gagal menurunkan stok semua varian barang induk ID %d: %v", data.IdBarangInduk, err_updates)
					return err_updates
				}
				return nil
			}); err != nil {
				log.Printf("[ERROR] Gagal downkan semua stok barang induk ID %d: %v", data.IdBarangInduk, err)
			}
		}()
	}

	log.Printf("[INFO] Semua stok barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseDownBarang{
			Message: "Berhasil",
		},
	}
}

func DownKategoriBarang(db *gorm.DB, data PayloadDownKategoriBarang) *response.ResponseForm {
	services := "DownKategoriBarang"

	var count_barang int64
	if err_db := db.Model(models.BarangInduk{}).Where(models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdentitasSeller.IdSeller}).Count(&count_barang).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdentitasSeller.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, barang tidak ditemukan.",
			},
		}
	} else {
		go func() {
			if err_db := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).
					Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: "Ready"}).
					Or(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: "Terjual"}).
					Update("status", "Down").Error; err_updates != nil {
					log.Printf("[ERROR] Gagal menurunkan stok kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, data.IdBarangInduk, err_updates)
					return err_updates
				}
				return nil
			}); err_db != nil {
				log.Printf("[ERROR] Gagal downkan stok kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, data.IdBarangInduk, err_db)
			}
		}()
	}

	log.Printf("[INFO] Semua stok kategori ID %d pada barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdKategoriBarang, data.IdBarangInduk, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseDownKategori{
			Message: "Berhasil",
		},
	}
}

func EditAlamatGudangBarangInduk(data PayloadEditAlamatBarangInduk, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudangBarangInduk"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit alamat gudang. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, barang induk tidak ditemukan atau kredensial seller tidak valid.",
			},
		}
	}

	if err_edit := db.Model(models.KategoriBarang{}).Where(models.KategoriBarang{
		IdBarangInduk: valid_id,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		log.Printf("[ERROR] Gagal update alamat gudang untuk semua kategori barang induk ID %d: %v", valid_id, err_edit)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil diubah untuk semua kategori barang induk ID %d oleh seller ID %d", valid_id, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditAlamatBarangInduk{
			Message: "Alamat gudang berhasil diubah.",
		},
	}
}

func EditAlamatGudangBarangKategori(data PayloadEditAlamatBarangKategori, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudangBarangKategori"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, kredensial seller tidak valid.",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit alamat gudang kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, barang induk tidak ditemukan atau kredensial seller tidak valid.",
			},
		}
	}

	if err_edit := db.Model(models.KategoriBarang{}).Where(models.KategoriBarang{
		IdBarangInduk: valid_id,
		ID:            data.IdKategoriBarang,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		log.Printf("[ERROR] Gagal update alamat gudang untuk kategori ID %d pada barang induk ID %d: %v", data.IdKategoriBarang, valid_id, err_edit)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil diubah untuk kategori ID %d pada barang induk ID %d oleh seller ID %d", data.IdKategoriBarang, valid_id, data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditAlamatBarangKategori{
			Message: "Alamat gudang berhasil diubah.",
		},
	}
}
