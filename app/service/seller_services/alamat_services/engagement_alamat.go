package seller_alamat_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_alamat_services_seller "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services/response_alamat_service_seller"
)

func TambahAlamatGudang(data PayloadTambahAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Gagal, kredensial tidak valid",
			},
		}
	}

	//
	data.Data.ID = 0
	//

	if err_tambah_alamat := db.Create(&data.Data).Error; err_tambah_alamat != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
			Message: "Berhasil",
		},
	}
}

func EditAlamatGudang(data PayloadEditAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "EditAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Gagal, kredensial tidak valid",
			},
		}
	}

	if err_edit_alamat := db.Model(models.AlamatGudang{}).Where(models.AlamatGudang{
		ID: data.Data.ID,
	}).Updates(&data.Data).Error; err_edit_alamat != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Gagal, Server sedang dibuk coba lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
			Message: "Berhasil",
		},
	}
}

func HapusAlamatGudang(data PayloadHapusAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal, kredensial tidak valid",
			},
		}
	}

	if err_hapus := db.Model(&models.AlamatGudang{}).Where(models.AlamatGudang{
		ID:       data.IdGudang,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Delete(&models.AlamatGudang{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
			Message: "Berhasil",
		},
	}
}
