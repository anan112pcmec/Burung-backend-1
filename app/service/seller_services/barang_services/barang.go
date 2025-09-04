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

func MasukanBarang(db *gorm.DB, data PayloadMasukanBarang) *response.ResponseForm {
	services := "MasukanBarang"
	data.BarangInduk.SellerID = data.IdSeller

	// Validasi data
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

	// jika err != nil, berarti tidak ada data, lanjut proses insert/update

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

	// Proses insert di background
	go func(data PayloadMasukanBarang) {
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
	}(data)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseMasukanBarang{
			Message: "Berhasil, Data Barang sedang diproses. Jika dalam beberapa menit data belum masuk, harap masukkan ulang karena server mungkin sibuk.",
		},
	}
}

func HapusBarang(db *gorm.DB, data PayloadHapusBarang) *response.ResponseForm {
	services := "HapusBarang"
	var barangdihapus int32

	// validasi dulu apakah barang dengan seller dan nama sesuai ada
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

	// eksekusi hapus dengan transaction
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			// hapus varian barang
			if err := tx.Unscoped().Where("id_barang_induk = ?", barangdihapus).
				Delete(&models.VarianBarang{}).Error; err != nil {
				return err
			}

			// hapus kategori barang
			if err := tx.Unscoped().Where("id_barang_induk = ?", barangdihapus).
				Delete(&models.KategoriBarang{}).Error; err != nil {
				return err
			}

			// hapus barang induk
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
			Message: fmt.Sprintf("Berhasil Menghapus Barang %s, Jika Barang Tidak Terhapus Coba Hapus Ulang", data.BarangInduk.NamaBarang),
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
	go func(copyData PayloadEditBarang) {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.BarangInduk{}).
				Where("id = ?", id).
				Updates(&copyData.BarangInduk).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("[EditBarang] Gagal update BarangInduk ID %d: %v", copyData.BarangInduk.ID, err)
		}
	}(data)

	// User langsung dapat respon tanpa nunggu goroutine selesai
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarang{
			Message: "Berhasil, Data Barang sedang diproses untuk diperbarui",
		},
	}
}
