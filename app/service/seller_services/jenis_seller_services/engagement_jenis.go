package jenis_seller_services

import (
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/jenis_seller_services/response_jenis_seller"
)

func AjukanUbahJenisSeller(data PayloadAjukanUbahJenisSeller, db *gorm.DB) *response.ResponseForm {
	services := "AjukanUbahJenisSeller"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		log.Printf("[WARN] Identitas seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, identitas seller tidak valid.",
			},
		}
	}

	var seller models.Seller
	if err := db.Model(&models.Seller{}).Where(models.Seller{
		ID:       data.DataDiajukan.IdSeller,
		Username: data.IdentitasSeller.Username,
		Email:    data.IdentitasSeller.EmailSeller,
	}).Limit(1).Take(&seller).Error; err != nil {
		log.Printf("[ERROR] Gagal mengambil data seller ID %d: %v", data.DataDiajukan.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	if seller.ID == 0 {
		log.Printf("[WARN] Data identitas tidak valid untuk seller ID %d", data.DataDiajukan.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, data identitas tidak valid.",
			},
		}
	}

	if seller.Jenis == data.DataDiajukan.TargetJenis {
		log.Printf("[WARN] Seller ID %d sudah memiliki jenis %s", seller.ID, seller.Jenis)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, kamu sudah berjenis demikian. Coba jenis lain.",
			},
		}
	}

	data.DataDiajukan.AlasanAdmin = ""
	data.DataDiajukan.ValidationStatus = "Pending"

	if err := db.Create(&data.DataDiajukan).Error; err != nil {
		log.Printf("[ERROR] Gagal mengajukan perubahan jenis seller ID %d: %v", data.DataDiajukan.IdSeller, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, server sedang sibuk. Coba lagi nanti.",
			},
		}
	}

	log.Printf("[INFO] Pengajuan perubahan jenis seller berhasil untuk seller ID %d", data.DataDiajukan.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
			Message: "Berhasil diajukan.",
		},
	}
}
