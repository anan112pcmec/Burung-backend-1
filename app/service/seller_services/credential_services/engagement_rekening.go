package seller_credential_services

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services/response_credential_seller"
)

func TambahRekeningSeller(data PayloadTambahkanNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "TambahRekeningSeller"

	if data.Data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var id_seller int64
	if check_seller := db.Model(&models.Seller{}).
		Select("id").
		Where(models.Seller{ID: data.Data.IDSeller}).
		First(&id_seller).Error; check_seller != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_seller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	// Transaction mulai dari cek rekening sampai insert
	err := db.Transaction(func(tx *gorm.DB) error {
		var id_rekening int64
		if err_check_rekening := tx.Model(&models.RekeningSeller{}).
			Select("id").
			Where(models.RekeningSeller{
				IDSeller:      data.Data.IDSeller,
				NamaBank:      data.Data.NamaBank,
				NomorRekening: data.Data.NomorRekening,
			}).
			First(&id_rekening).Error; err_check_rekening == nil {
			// sudah ada rekening
			return fmt.Errorf("rekening sudah ada")
		}

		// Insert data rekening baru
		if err_masukan := tx.Create(&data.Data).Error; err_masukan != nil {
			return err_masukan
		}

		return nil
	})

	if err != nil {
		if err.Error() == "rekening sudah ada" {
			return &response.ResponseForm{
				Status:   http.StatusConflict,
				Services: services,
				Payload: response_credential_seller.ResponseTambahRekeningSeller{
					Message: "Data Rekening itu sudah ada dan tercatat di akun mu",
				},
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseTambahRekeningSeller{
			Message: "Berhasil",
		},
	}
}

func HapusRekeningSeller(data PayloadHapusNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "HapusRekeningSeller"

	if data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var id_rekening int64

		if check_rekening := tx.Model(&models.RekeningSeller{}).
			Select("id").
			Where(models.RekeningSeller{
				IDSeller:        data.IDSeller,
				NamaBank:        data.NamaBank,
				PemilikRekening: data.PemilikRekening,
				NomorRekening:   data.NomorRekening,
			}).
			First(&id_rekening).Error; check_rekening != nil {

			return check_rekening
		}

		if hapus_rekening := tx.Model(&models.RekeningSeller{}).
			Where(models.RekeningSeller{ID: id_rekening}).
			Delete(&models.RekeningSeller{}).Error; hapus_rekening != nil {

			return hapus_rekening
		}

		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseHapusRekeningSeller{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseHapusRekeningSeller{
			Message: "Berhasil",
		},
	}
}
