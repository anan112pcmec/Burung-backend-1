package seller_service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services/response_barang_service"

)

// ////////////////////////////////////////////////////////////////////////////////
// BARANG INDUK
// ////////////////////////////////////////////////////////////////////////////////

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
		ctx := context.Background()
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
						Status:        models.Ready,
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

	// User langsung dapat respon tanpa nunggu goroutine selesai
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
						Status:        models.Ready,
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

	go func(data PayloadHapusKategori) {
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

				if err := tx.Unscoped().Where(&models.VarianBarang{
					IdKategori:    existingKategori.ID,
					IdBarangInduk: data.IdBarangInduk,
				}).Delete(&models.VarianBarang{}).Error; err != nil {
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
	}(data)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseHapusKategori{
			Message: "Berhasil",
		},
	}
}

func EditKategoriBarang(db *gorm.DB, data PayloadEditKategori) *response.ResponseForm {
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

	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Unscoped().Updates(&data.KategoriBarang).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[TambahKategoriBarang] Gagal menghapus kategori untuk BarangInduk ID %s: %v", idBarangIndukGet, err)
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
