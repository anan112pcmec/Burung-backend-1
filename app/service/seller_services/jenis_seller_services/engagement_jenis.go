package jenis_seller_services

import (
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
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, Identitas seller tidak ditemukan",
			},
		}
	}

	var seller models.Seller
	if err := db.Model(&models.Seller{}).Where(models.Seller{
		ID:       data.DataDiajukan.IdSeller,
		Username: data.IdentitasSeller.Username,
		Email:    data.IdentitasSeller.EmailSeller,
	}).Limit(1).Take(&seller).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, server sedang sibuk, coba lagi nanti",
			},
		}
	}

	if seller.ID == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, data identitas tidak valid",
			},
		}
	}

	if seller.Jenis == data.DataDiajukan.TargetJenis {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, kamu sudah berjenis demikian, coba jenis lain",
			},
		}
	}

	data.DataDiajukan.AlasanAdmin = ""
	data.DataDiajukan.ValidationStatus = "Pending"

	if err := db.Create(&data.DataDiajukan).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
				Message: "Gagal, server sedang sibuk, coba lagi nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_jenis_seller.ResponseAjukanUbahJenisSeller{
			Message: "Berhasil diajukan",
		},
	}
}
