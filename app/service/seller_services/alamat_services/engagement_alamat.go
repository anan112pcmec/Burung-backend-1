package seller_alamat_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_alamat_services_seller "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services/response_alamat_service_seller"
)

func MasukanAlamatSeller(data PayloadMasukanAlamatSeller, db *gorm.DB) *response.ResponseForm {
	services := "MasukanAlamatSeller"

	if data.Data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var jumlah int64
	if err_lebih := db.Model(&models.AlamatSeller{}).Where("id_seller = ?", data.Data.IDSeller).Count(&jumlah).Error; err_lebih != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
				Messages: "Gagal, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	if jumlah >= 5 {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
				Messages: "Gagal, Alamat Mu sudah mencapai batas",
			},
		}
	}

	var existing models.AlamatSeller
	errcheck := db.Where(&models.AlamatSeller{
		IDSeller:        data.Data.IDSeller,
		PanggilanAlamat: data.Data.PanggilanAlamat,
	}).First(&existing).Error

	if errcheck != nil && errcheck != gorm.ErrRecordNotFound {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
				Messages: "Gagal, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	if errcheck == gorm.ErrRecordNotFound {
		if err := db.Create(&models.AlamatSeller{
			IDSeller:        data.Data.IDSeller,
			PanggilanAlamat: data.Data.PanggilanAlamat,
			NamaAlamat:      data.Data.NamaAlamat,
			Deskripsi:       data.Data.Deskripsi,
			NomorTelephone:  data.Data.NomorTelephone,
			Longitude:       data.Data.Longitude,
			Latitude:        data.Data.Latitude,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
					Messages: "Gagal server sedang sibuk coba lagi nanti",
				},
			}
		}
	} else {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
				Messages: "Gagal Alamat Dengan Panggilan itu sudah ada ganti panggilan nya",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseMasukanAlamatSeller{
			Messages: "Berhasil",
		},
	}
}

func HapusAlamatSeller(data PayloadHapusAlamatSeller, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatSeller"

	if data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatSeller{
				Messages: "ID Seller tidak valid",
			},
		}
	}

	// Hapus langsung
	result := db.Where(&models.AlamatSeller{
		IDSeller:        data.IDSeller,
		PanggilanAlamat: data.PanggilanAlamat,
	}).Delete(&models.AlamatSeller{})

	// Cek error query
	if result.Error != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatSeller{
				Messages: "Gagal, coba lagi nanti. Server sedang sibuk",
			},
		}
	}

	// Kalau tidak ada data yang terhapus
	if result.RowsAffected == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatSeller{
				Messages: "Alamat tidak ditemukan atau sudah dihapus",
			},
		}
	}

	// Kalau berhasil
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseHapusAlamatSeller{
			Messages: "Alamat berhasil dihapus",
		},
	}
}
