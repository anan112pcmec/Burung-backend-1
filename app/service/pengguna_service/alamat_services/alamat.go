package pengguna_alamat_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/alamat_services/response_alamat_service_pengguna"
)

func MasukanAlamatPengguna(data PayloadMasukanAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "MasukanAlamatPengguna"

	if data.DataAlamat.IDPengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var count int64
	db.Model(&models.AlamatPengguna{}).
		Where(models.AlamatPengguna{IDPengguna: data.DataAlamat.IDPengguna}).
		Count(&count)

	if count >= 5 {
		return &response.ResponseForm{
			Status:   http.StatusForbidden,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
				Messages: "Gagal, maksimal hanya boleh menyimpan 5 alamat",
			},
		}
	}

	var existing models.AlamatPengguna
	errcheck := db.Where(models.AlamatPengguna{
		IDPengguna:      data.DataAlamat.IDPengguna,
		PanggilanAlamat: data.DataAlamat.PanggilanAlamat,
	}).First(&existing).Error

	if errcheck == gorm.ErrRecordNotFound {
		if err := db.Create(&models.AlamatPengguna{
			IDPengguna:      data.DataAlamat.IDPengguna,
			PanggilanAlamat: data.DataAlamat.PanggilanAlamat,
			NamaAlamat:      data.DataAlamat.NamaAlamat,
			Deskripsi:       data.DataAlamat.Deskripsi,
			NomorTelephone:  data.DataAlamat.NomorTelephone,
			Longitude:       data.DataAlamat.Longitude,
			Latitude:        data.DataAlamat.Latitude,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
					Messages: "Gagal server sedang sibuk coba lagi nanti",
				},
			}
		}
	} else {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
				Messages: "Gagal Alamat Dengan Panggilan itu sudah ada ganti panggilan nya",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
			Messages: "Berhasil",
		},
	}
}

func HapusAlamatPengguna(data PayloadHapusAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatPengguna"

	if data.IdPengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err_hapus := db.Where(models.AlamatPengguna{
		IDPengguna:      data.IdPengguna,
		PanggilanAlamat: data.PanggilanAlamat,
	}).Delete(&models.AlamatPengguna{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamat{
				Messages: "Gagal hapus server sedang sibuk, Coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseHapusAlamat{
			Messages: "Berhasil",
		},
	}
}
