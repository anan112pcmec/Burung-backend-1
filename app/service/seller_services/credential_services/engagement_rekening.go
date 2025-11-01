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

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Tambahkan Rekening Seller
// Berfungsi untuk menambahkan rekening seller ke database
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahRekeningSeller(data PayloadTambahkanNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "TambahRekeningSeller"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_seller.ResponseTambahRekeningSeller{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var id_rekening int64
		if err_check_rekening := tx.Model(&models.RekeningSeller{}).
			Select("id").
			Where(models.RekeningSeller{
				IDSeller:      data.IdentitasSeller.IdSeller,
				NamaBank:      data.Data.NamaBank,
				NomorRekening: data.Data.NomorRekening,
			}).
			First(&id_rekening).Error; err_check_rekening == nil {
			return fmt.Errorf("rekening sudah ada")
		}

		var hitung int64 = 0

		if err := tx.Model(&models.RekeningSeller{}).Where(&models.RekeningSeller{
			IDSeller: data.IdentitasSeller.IdSeller,
		}).Count(&hitung).Error; err != nil {
			return err
		}
		if hitung == 0 {
			data.Data.IsDefault = true
		} else {
			data.Data.IsDefault = false
		}

		data.Data.IDSeller = data.IdentitasSeller.IdSeller

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

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Rekening Seller
// Berfungsi untuk Menghapus Data Rekening Seller Yang sudah ada di db
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusRekeningSeller(data PayloadHapusNorekSeller, db *gorm.DB) *response.ResponseForm {
	services := "HapusRekeningSeller"

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_seller.ResponseHapusRekeningSeller{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		var id_rekening int64

		if check_rekening := tx.Model(&models.RekeningSeller{}).
			Select("id").
			Where(models.RekeningSeller{
				IDSeller:        data.IdentitasSeller.IdSeller,
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
		log.Printf("[ERROR] Gagal menghapus rekening untuk seller ID %d: %v", data.IdentitasSeller.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseHapusRekeningSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Rekening berhasil dihapus untuk seller ID %d", data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseHapusRekeningSeller{
			Message: "Rekening berhasil dihapus.",
		},
	}
}
