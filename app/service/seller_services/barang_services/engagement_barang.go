package seller_service

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services/response_barang_service"
)

func MasukanBarang(db *gorm.DB, data PayloadMasukanBarang) *response.ResponseForm {
	services := "MasukanBarang"
	data.BarangInduk.SellerID = data.IdSeller

	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		log.Printf("[WARN] Data barang tidak lengkap untuk seller ID %d", data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal: Data Barang Harus Dilengkapi",
		}
	}

	var namabarang string
	errnama := db.Model(&models.BarangInduk{}).
		Where("id_seller = ? AND nama_barang = ?", data.IdSeller, data.BarangInduk.NamaBarang).
		Select("nama_barang").
		Take(&namabarang).Error

	if errnama == nil {
		log.Printf("[WARN] Seller ID %d sudah memiliki barang dengan nama '%s'", data.IdSeller, data.BarangInduk.NamaBarang)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  "Gagal: Nama barang sudah terdaftar untuk seller ini",
		}
	}

	var sellerID int32
	err := db.Model(&models.Seller{}).
		Select("id").
		Where("id = ?", data.IdSeller).
		Take(&sellerID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[WARN] Seller ID %d tidak ditemukan", data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal: Seller tidak ditemukan",
		}
	}
	if err != nil {
		log.Printf("[ERROR] Terjadi kesalahan saat validasi seller ID %d: %v", data.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal: Terjadi kesalahan saat validasi seller",
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			if data.BarangInduk.OriginalKategori == "" {
				return fmt.Errorf("OriginalKategori kosong, rollback transaksi")
			}

			if err := tx.Model(&models.BarangInduk{}).Create(&data.BarangInduk).Error; err != nil {
				return err
			}

			idInduk := data.BarangInduk.ID

			for _, kategori := range data.KategoriBarang {
				kategori.IdBarangInduk = int32(idInduk)

				if err := tx.Model(&models.KategoriBarang{}).Create(&kategori).Error; err != nil {
					return err
				}

				idKategori := kategori.ID

				var varianBatch []models.VarianBarang
				for i := 0; i < int(kategori.Stok); i++ {
					varianBatch = append(varianBatch, models.VarianBarang{
						IdBarangInduk: int32(idInduk),
						IdKategori:    idKategori,
						Status:        "Ready",
						Sku:           kategori.Sku,
					})
				}

				if len(varianBatch) > 0 {
					if err := tx.Model(&models.VarianBarang{}).Create(&varianBatch).Error; err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] MasukanBarang gagal insert data untuk seller ID %d: %v", data.IdSeller, err)
		}
	}()

	log.Printf("[INFO] Barang berhasil dimasukkan untuk seller ID %d", data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseMasukanBarang{
			Message: "Barang berhasil ditambahkan.",
		},
	}
}

func HapusBarang(db *gorm.DB, data PayloadHapusBarang) *response.ResponseForm {
	services := "HapusBarang"
	var barangdihapus int32

	if err := db.Model(&models.BarangInduk{}).
		Where("id_seller = ? AND nama_barang = ?", data.IdSeller, data.BarangInduk.NamaBarang).
		Select("id").
		Take(&barangdihapus).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[WARN] Barang tidak ditemukan untuk seller ID %d dengan nama '%s'", data.IdSeller, data.BarangInduk.NamaBarang)
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  "Barang tidak ditemukan atau tidak sesuai kredensial seller",
			}
		}
		log.Printf("[ERROR] Terjadi kesalahan pada database saat hapus barang: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	var jumlah_dalam_transaksi int64
	if err_stock := db.Model(models.VarianBarang{}).Where(models.VarianBarang{IdBarangInduk: barangdihapus, Status: "Dipesan"}).Or(models.VarianBarang{Status: "Diproses"}).Count(&jumlah_dalam_transaksi).Error; err_stock != nil {
		log.Printf("[ERROR] Gagal cek stok dalam transaksi saat hapus barang: %v", err_stock)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	if jumlah_dalam_transaksi != 0 {
		log.Printf("[WARN] Masih ada %v barang dalam transaksi untuk barang ID %d", jumlah_dalam_transaksi, barangdihapus)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarang{
				Message: fmt.Sprintf("Masih ada sejumlah: %v barang dalam transaksi", jumlah_dalam_transaksi),
			},
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Unscoped().Where("id_barang_induk = ?", barangdihapus).
				Delete(&models.VarianBarang{}).Error; err != nil {
				return err
			}

			if err := tx.Unscoped().Where("id_barang_induk = ?", barangdihapus).
				Delete(&models.KategoriBarang{}).Error; err != nil {
				return err
			}

			if err := tx.Unscoped().Where("id = ?", barangdihapus).
				Delete(&models.BarangInduk{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			log.Printf("[ERROR] Rollback HapusBarang untuk barang ID %d: %v", barangdihapus, err)
		}
	}()

	log.Printf("[INFO] Barang berhasil dihapus untuk seller ID %d", data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusBarang{
			Message: "Barang berhasil dihapus.",
		},
	}
}

func EditBarang(db *gorm.DB, data PayloadEditBarang) *response.ResponseForm {
	services := "EditBarang"
	data.BarangInduk.SellerID = data.IdSeller

	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		log.Printf("[WARN] Data barang tidak lengkap untuk edit seller ID %d", data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Gagal mengubah data barang: Data tidak lengkap",
			},
		}
	}

	var id int32
	err := db.Model(&models.BarangInduk{}).
		Where("nama_barang = ? AND id_seller = ?", data.BarangInduk.NamaBarang, data.IdSeller).
		Select("id").
		Take(&id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[WARN] Barang tidak valid untuk edit seller ID %d", data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Barang tidak valid",
			},
		}
	} else if err != nil {
		log.Printf("[ERROR] Terjadi kesalahan server saat validasi barang: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Terjadi kesalahan server saat validasi barang",
			},
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.BarangInduk{}).
				Where("id = ?", id).
				Updates(&data.BarangInduk).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditBarang gagal update BarangInduk ID %d: %v", data.BarangInduk.ID, err)
		}
	}()

	log.Printf("[INFO] Barang berhasil diubah untuk seller ID %d", data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarang{
			Message: "Barang berhasil diubah.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// KATEGORI BARANG
// ////////////////////////////////////////////////////////////////////////////////

func TambahKategoriBarang(db *gorm.DB, data PayloadTambahKategori) *response.ResponseForm {
	services := "TambahKategoriBarang"

	var idBarangIndukGet string
	barangIndukFilter := models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdSeller,
	}

	if err := db.Model(&models.BarangInduk{}).
		Where(&barangIndukFilter).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk tambah kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	go func(data PayloadTambahKategori) {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				data.KategoriBarang[i].IdBarangInduk = data.IdBarangInduk

				// cek apakah kategori sudah ada
				var existingKategori models.KategoriBarang
				err := tx.Where("nama = ? AND id_barang_induk = ?", data.KategoriBarang[i].Nama, data.IdBarangInduk).
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
				if err := tx.Where("nama = ? AND id_barang_induk = ?", data.KategoriBarang[i].Nama, data.IdBarangInduk).
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

	log.Printf("[INFO] Kategori barang berhasil ditambahkan pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseTambahKategori{
			Message: "Kategori barang berhasil ditambahkan.",
		},
	}
}

func HapusKategoriBarang(db *gorm.DB, data PayloadHapusKategori) *response.ResponseForm {
	services := "HapusKategoriBarang"

	var idBarangIndukGet string
	barangIndukFilter := models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdSeller,
	}

	if err := db.Model(&models.BarangInduk{}).
		Where(&barangIndukFilter).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk hapus kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
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

				if err := tx.Unscoped().Delete(&existingKategori).Error; err != nil {
					log.Printf("[ERROR] Gagal hapus kategori ID %d: %v", existingKategori.ID, err)
					return err
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] HapusKategoriBarang gagal proses hapus kategori & varian untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil dihapus pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusKategori{
			Message: "Kategori barang berhasil dihapus.",
		},
	}
}

func EditKategoriBarang(db *gorm.DB, data PayloadEditKategori) *response.ResponseForm {
	services := "EditKategoriBarang"

	var idBarangIndukGet string
	barangIndukFilter := models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdSeller,
	}

	if err := db.Model(&models.BarangInduk{}).
		Where(&barangIndukFilter).
		Select("id").
		First(&idBarangIndukGet).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk edit kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal mendapatkan barang yang dituju",
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				// update kategori barang
				if err := tx.Model(&models.KategoriBarang{}).
					Where("id = ?", data.KategoriBarang[i].ID).
					Updates(&data.KategoriBarang[i]).Error; err != nil {
					log.Printf("[ERROR] Gagal update kategori ID %d: %v", data.KategoriBarang[i].ID, err)
					return err
				}

				varBarangFilter := models.VarianBarang{
					IdBarangInduk: data.IdBarangInduk,
					IdKategori:    data.KategoriBarang[i].ID,
				}
				if err := tx.Model(&models.VarianBarang{}).
					Where(&varBarangFilter).
					Update("sku", data.KategoriBarang[i].Sku).Error; err != nil {
					log.Printf("[ERROR] Gagal update SKU varian untuk kategori ID %d: %v", data.KategoriBarang[i].ID, err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("[ERROR] EditKategoriBarang gagal update kategori untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	log.Printf("[INFO] Kategori barang berhasil diubah pada barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditKategori{
			Message: "Kategori barang berhasil diubah.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////
// STOK BARANG
// ////////////////////////////////////////////////////////////////////////////////

func EditStokBarang(db *gorm.DB, data PayloadEditStokBarang) *response.ResponseForm {
	services := "EditStokBarang"

	var idbaranginduk models.BarangInduk
	if err := db.
		Where(&models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdSeller}).
		Select("id").
		Take(&idbaranginduk).Error; err != nil {
		log.Printf("[WARN] Barang induk tidak ditemukan: IdBarangInduk=%d, IdSeller=%d, Error=%v", data.IdBarangInduk, data.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Barang induk tidak ditemukan / tidak sesuai seller",
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
						Where("id_barang_induk = ? AND id_kategori = ? AND status = ?", data.IdBarangInduk, b.IdKategoriBarang, "Ready").
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

	log.Printf("[INFO] Proses update stok sedang berjalan untuk barang induk ID %d oleh seller ID %d", data.IdBarangInduk, data.IdSeller)
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

	var count_barang int64
	if err_db := db.Model(models.BarangInduk{}).Where(models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdSeller}).Count(&count_barang).Limit(1).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdSeller)
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

	log.Printf("[INFO] Semua stok barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdBarangInduk, data.IdSeller)
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
	if err_db := db.Model(models.BarangInduk{}).Where(models.BarangInduk{ID: data.IdBarangInduk, SellerID: data.IdSeller}).Count(&count_barang).Error; err_db != nil {
		log.Printf("[ERROR] Gagal mengambil data barang induk ID %d untuk seller ID %d: %v", data.IdBarangInduk, data.IdSeller, err_db)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if count_barang != 1 {
		log.Printf("[WARN] Barang induk tidak ditemukan untuk down stok kategori. IdBarangInduk=%d, IdSeller=%d", data.IdBarangInduk, data.IdSeller)
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

	log.Printf("[INFO] Semua stok kategori ID %d pada barang induk ID %d berhasil di-down-kan oleh seller ID %d", data.IdKategoriBarang, data.IdBarangInduk, data.IdSeller)
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
