package seller_credential_services

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services/response_credential_seller"
)

func TambahRekeningSeller(data PayloadTambahkanNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "TambahRekeningSeller"

	if data.Data.IDSeller == 0 {
		log.Println("[WARN] ID seller tidak ditemukan pada permintaan tambah rekening.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "ID seller tidak ditemukan.",
			},
		}
	}

	var id_seller int64
	if check_seller := db.Model(&models.Seller{}).
		Select("id").
		Where(models.Seller{ID: data.Data.IDSeller}).
		First(&id_seller).Error; check_seller != nil {
		log.Printf("[ERROR] Gagal validasi seller ID %d: %v", data.Data.IDSeller, check_seller)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	if id_seller == 0 {
		log.Printf("[WARN] Seller ID %d tidak ditemukan saat tambah rekening.", data.Data.IDSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Seller tidak ditemukan.",
			},
		}
	}

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
			return fmt.Errorf("rekening sudah ada")
		}

		if err_masukan := tx.Create(&data.Data).Error; err_masukan != nil {
			return err_masukan
		}

		return nil
	})

	if err != nil {
		if err.Error() == "rekening sudah ada" {
			log.Printf("[WARN] Rekening sudah ada untuk seller ID %d: %s - %s", data.Data.IDSeller, data.Data.NamaBank, data.Data.NomorRekening)
			return &response.ResponseForm{
				Status:   http.StatusConflict,
				Services: services,
				Payload: response_credential_seller.ResponseTambahRekeningSeller{
					Message: "Data rekening tersebut sudah ada dan tercatat di akun Anda.",
				},
			}
		}
		log.Printf("[ERROR] Gagal menambah rekening untuk seller ID %d: %v", data.Data.IDSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Rekening berhasil ditambahkan untuk seller ID %d", data.Data.IDSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseTambahRekeningSeller{
			Message: "Rekening berhasil ditambahkan.",
		},
	}
}

func HapusRekeningSeller(data PayloadHapusNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "HapusRekeningSeller"

	if data.IDSeller == 0 {
		log.Println("[WARN] ID seller tidak ditemukan pada permintaan hapus rekening.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_seller.ResponseHapusRekeningSeller{
				Message: "ID seller tidak ditemukan.",
			},
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
		log.Printf("[ERROR] Gagal menghapus rekening untuk seller ID %d: %v", data.IDSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseHapusRekeningSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Rekening berhasil dihapus untuk seller ID %d", data.IDSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseHapusRekeningSeller{
			Message: "Rekening berhasil dihapus.",
		},
	}
}
