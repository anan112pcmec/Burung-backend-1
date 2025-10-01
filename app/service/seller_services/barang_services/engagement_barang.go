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
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  "Gagal: Oops Kamu Coba Ganti Nama Barang, Kamu Sudah Punya Barang Itu",
		}
	}

	fmt.Println("errnama udh ada tapi gajalan")

	var sellerID int32
	err := db.Model(&models.Seller{}).
		Select("id").
		Where("id = ?", data.IdSeller).
		Take(&sellerID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal: Seller tidak ditemukan",
		}
	}
	if err != nil {
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
			log.Printf("[MasukanBarang] gagal insert data: %v", err)
		}
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseMasukanBarang{
			Message: "Berhasil",
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
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  "Barang tidak ditemukan atau tidak sesuai kredensial seller",
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusBadGateway,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	var jumlah_dalam_transaksi int64
	if err_stock := db.Model(models.VarianBarang{}).Where(models.VarianBarang{IdBarangInduk: barangdihapus, Status: "Dipesan"}).Or(models.VarianBarang{Status: "Diproses"}).Count(&jumlah_dalam_transaksi).Error; err_stock != nil {
		return &response.ResponseForm{
			Status:   http.StatusBadGateway,
			Services: services,
			Payload:  "Terjadi kesalahan pada database",
		}
	}

	if jumlah_dalam_transaksi != 0 {
		return &response.ResponseForm{
			Status:   http.StatusBadGateway,
			Services: services,
			Payload: response_barang_service.ResponseHapusBarang{
				Message: fmt.Sprintf("Masih Ada Sejumlah: %v Barang Dalam Transaksi", jumlah_dalam_transaksi),
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
			fmt.Println("Rollback HapusBarang:", err)
		}
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusBarang{
			Message: "Berhasil",
		},
	}
}

func EditBarang(db *gorm.DB, data PayloadEditBarang) *response.ResponseForm {
	services := "EditBarang"
	data.BarangInduk.SellerID = data.IdSeller

	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Gagal Mengubah Data Barang: Data Tidak Lengkap",
			},
		}
	}

	var id int32
	err := db.Model(&models.BarangInduk{}).
		Where("nama_barang = ? AND id_seller = ?", data.BarangInduk.NamaBarang, data.IdSeller).
		Select("id").
		Take(&id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Barang Tidak Valid",
			},
		}
	} else if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Terjadi kesalahan server saat validasi Barang",
			},
		}
	}

	// Async update (fire-and-forget)
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.BarangInduk{}).
				Where("id = ?", id).
				Updates(&data.BarangInduk).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[EditBarang] Gagal update BarangInduk ID %d: %v", data.BarangInduk.ID, err)
		}
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarang{
			Message: "Berhasil",
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
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Gagal Mendapatkan Barang Yang Dituju",
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
					continue
				} else if err != gorm.ErrRecordNotFound {
					log.Printf("[TambahKategoriBarang] Error cek kategori %s: %v", data.KategoriBarang[i].Nama, err)
					continue
				}

				if err := tx.Create(&data.KategoriBarang[i]).Error; err != nil {
					return err
				}

				var kategoriBaru models.KategoriBarang
				if err := tx.Where("nama = ? AND id_barang_induk = ?", data.KategoriBarang[i].Nama, data.IdBarangInduk).
					First(&kategoriBaru).Error; err != nil {
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
						return err
					}
				}
			}

			return nil
		}); err != nil {
			log.Printf("[TambahKategoriBarang] Gagal proses kategori & varian untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}(data)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseTambahKategori{
			Message: "Berhasil",
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
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload:  "Gagal Mendapatkan Barang Yang Dituju",
		}
	}

	for _, kat := range data.KategoriBarang {
		var barang_dalam_transaksi int64
		if err_stock := db.Model(models.VarianBarang{}).Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: kat.ID, Status: "Dipesan"}).Or(models.VarianBarang{Status: "Diproses"}).Count(&barang_dalam_transaksi).Error; err_stock != nil {
			continue
		}

		if barang_dalam_transaksi != 0 {
			return &response.ResponseForm{
				Status:   http.StatusOK,
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
						continue
					} else {
						log.Printf("[HapusKategoriBarang] Error cek kategori ID %d: %v", data.KategoriBarang[i].ID, err)
						continue
					}
				}

				if err := tx.Unscoped().
					Where("id_kategori = ? AND id_barang_induk = ?", existingKategori.ID, data.IdBarangInduk).
					Delete(&models.VarianBarang{}).Error; err != nil {
					log.Printf("[HapusKategoriBarang] Gagal hapus varian untuk kategori ID %d: %v", existingKategori.ID, err)
					return err
				}

				if err := tx.Unscoped().Delete(&existingKategori).Error; err != nil {
					log.Printf("[HapusKategoriBarang] Gagal hapus kategori ID %d: %v", existingKategori.ID, err)
					return err
				}
			}

			return nil
		}); err != nil {
			log.Printf("[HapusKategoriBarang] Gagal hapus kategori & varian untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusKategori{
			Message: "Berhasil",
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
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload:  "Gagal Mendapatkan Barang Yang Dituju",
		}
	}

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			for i := range data.KategoriBarang {
				// update kategori barang
				if err := tx.Model(&models.KategoriBarang{}).
					Where("id = ?", data.KategoriBarang[i].ID).
					Updates(&data.KategoriBarang[i]).Error; err != nil {
					return err
				}

				varBarangFilter := models.VarianBarang{
					IdBarangInduk: data.IdBarangInduk,
					IdKategori:    data.KategoriBarang[i].ID,
				}
				if err := tx.Model(&models.VarianBarang{}).
					Where(&varBarangFilter).
					Update("sku", data.KategoriBarang[i].Sku).Error; err != nil {
					log.Printf("[EditKategoriBarang] Gagal update SKU: %v", err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("[EditKategoriBarang] Gagal mengupdate kategori untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
		}
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditKategori{
			Message: "Berhasil",
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
		log.Printf("[TRACE] Barang induk tidak ditemukan: IdBarangInduk=%d, IdSeller=%d, Error=%v\n", data.IdBarangInduk, data.IdSeller, err)
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
			log.Printf("[TRACE] Kategori tidak ditemukan: IdKategori=%d, NamaKategori=%s, Error=%v\n", b.IdKategoriBarang, b.NamaKategoriBarang, err)
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  fmt.Sprintf("Kategori %s tidak ditemukan", b.NamaKategoriBarang),
			}
		}
		log.Printf("[TRACE] Kategori ditemukan: IdKategori=%d, NamaKategori=%s\n", b.IdKategoriBarang, b.NamaKategoriBarang)
	}

	go func() {
		for _, b := range data.Barang {

			var jumlah int64
			if err := db.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: b.IdKategoriBarang, Status: "Ready"}).
				Count(&jumlah).Error; err != nil {
				log.Printf("[TRACE] Gagal menghitung stok: IdKategori=%d, Error=%v\n", b.IdKategoriBarang, err)
				continue
			}

			if jumlah == int64(b.JumlahStok) {
				return
			}

			if jumlah > int64(b.JumlahStok) {
				hapus := jumlah - int64(b.JumlahStok)
				log.Printf("[TRACE] Perlu hapus varian: IdKategori=%d, JumlahHapus=%d\n", b.IdKategoriBarang, hapus)
				for j := int64(0); j < hapus; j++ {
					if err := db.
						Where("id_barang_induk = ? AND id_kategori = ? AND status = ?", data.IdBarangInduk, b.IdKategoriBarang, "Ready").
						Limit(1).
						Delete(&models.VarianBarang{}).Error; err != nil {
						log.Printf("[TRACE] Gagal menghapus varian: IdKategori=%d, Iterasi=%d, Error=%v\n", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[TRACE] Berhasil menghapus varian: IdKategori=%d, Iterasi=%d\n", b.IdKategoriBarang, j)
					}
				}
			}

			if int64(b.JumlahStok) > jumlah {
				buat := int64(b.JumlahStok) - jumlah
				for j := int64(0); j < buat; j++ {
					newVarian := models.VarianBarang{
						IdBarangInduk: data.IdBarangInduk,
						IdKategori:    b.IdKategoriBarang,
						Sku:           b.SkuKategoriBarang,
					}
					if err := db.Create(&newVarian).Error; err != nil {
						log.Printf("[TRACE] Gagal membuat varian: IdKategori=%d, Iterasi=%d, Error=%v\n", b.IdKategoriBarang, j, err)
					} else {
						log.Printf("[TRACE] Berhasil membuat varian: IdKategori=%d, Iterasi=%d\n", b.IdKategoriBarang, j)
					}
				}
			}
		}
	}()

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
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal Coba Lagi Nanti",
			},
		}
	}

	if count_barang != 1 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownBarang{
				Message: "Gagal Barang Tidak Ditemukan, Coba Lagi Nanti",
			},
		}
	} else {
		go func() {
			if err := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, Status: "Ready"}).Or(models.VarianBarang{Status: "Terjual"}).Update("status", "Down").Error; err_updates != nil {
					return err_updates
				}
				return nil
			}); err != nil {
				log.Printf("Gagal Downkan Semua stok Barang")
			}
		}()
	}

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
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, Server Sedang sibuk coba lagi nanti",
			},
		}
	}

	if count_barang != 1 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseDownKategori{
				Message: "Gagal, Barang Tidak Ditemukan",
			},
		}
	} else {
		go func() {
			if err_db := db.Transaction(func(tx *gorm.DB) error {
				if err_updates := db.Model(models.VarianBarang{}).Where(models.VarianBarang{IdBarangInduk: data.IdBarangInduk, IdKategori: data.IdKategoriBarang, Status: "Ready"}).Or(models.VarianBarang{Status: "Terjual"}).Update("status", "Down").Error; err_updates != nil {
					return err_updates
				}
				return nil
			}); err_db != nil {
				log.Printf("Gagal Downkan stok kategori")
			}
		}()
	}

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
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, Kredensial Seller tidak valid",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, Kredensial Seller tidak valid",
			},
		}
	}

	if err_edit := db.Model(models.KategoriBarang{}).Where(models.KategoriBarang{
		IdBarangInduk: valid_id,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangInduk{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}
}

func EditAlamatGudangBarangKategori(data PayloadEditAlamatBarangKategori, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudangBarangKategori"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, Kredensial Seller tidak valid",
			},
		}
	}

	var valid_id int32 = 0
	_ = db.Model(models.BarangInduk{}).Select("id").Where(models.BarangInduk{
		ID:       data.IdBarangInduk,
		SellerID: data.IdentitasSeller.IdSeller,
	}).Take(&valid_id)

	if valid_id == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, Kredensial Seller tidak valid",
			},
		}
	}

	if err_edit := db.Model(models.KategoriBarang{}).Where(models.KategoriBarang{
		IdBarangInduk: valid_id,
		ID:            data.IdKategoriBarang,
	}).Update("id_alamat_gudang", data.IdAlamatGudang).Error; err_edit != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_barang_service.ResponseEditAlamatBarangKategori{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}
}
