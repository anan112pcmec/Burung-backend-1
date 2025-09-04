package seller_service

import (
	"context"
	"errors"
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

	go func(data PayloadMasukanBarang) {
		ctx := context.Background()
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

			if err := tx.Create(&data.BarangInduk).Error; err != nil {
				return err
			}

			for _, kategori := range data.KategoriBarang {
				kategori.IdBarangInduk = data.BarangInduk.ID
				if err := tx.Create(&kategori).Error; err != nil {
					return err
				}

				varianBarang := models.VarianBarang{
					IdBarangInduk: data.BarangInduk.ID,
					NamaKategori:  kategori.Nama,
					Status:        models.Ready,
					Sku:           kategori.Sku,
				}

				for range kategori.Stok {
					if err := tx.Create(&varianBarang).Error; err != nil {
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

func HapusBarang(db *gorm.DB, data PayloadHapusBarang) {

}

func EditBarang(db *gorm.DB, data PayloadEditBarang) *response.ResponseForm {
	services := "EditBarang"

	if validasi := data.BarangInduk.Validating(); validasi != "Data Lengkap" {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_barang_service.ResponseEditBarang{
				Message: "Gagal Mengubah Data Barang: Data Tidak Lengkap",
			},
		}
	}

	var id int64
	err := db.Where(models.BarangInduk{ID: data.BarangInduk.ID}, models.BarangInduk{SellerID: data.IdSeller}).
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

	go func(data PayloadEditBarang) {
		if err := db.Save(&data.BarangInduk).Error; err != nil {
			log.Printf("[EditBarang] Gagal update BarangInduk ID %d: %v", data.BarangInduk.ID, err)
		}
	}(data)

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_service.ResponseEditBarang{
			Message: "Berhasil, Data Barang sedang diproses untuk diperbarui",
		},
	}
}
