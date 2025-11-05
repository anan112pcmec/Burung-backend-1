package kurir_alamat_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/alamat_services/response_alamat_service_kurir"
)

func MasukanAlamatKurir(data PayloadMasukanAlamatKurir, db *gorm.DB) *response.ResponseForm {
	services := "MasukanAlamatKurir"

	// Validasi identitas kurir
	_, valid := data.IdentitasKurir.Validating(db)
	if !valid {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseMasukanAlamatKurir{
				Message: "Gagal: Data kurir tidak valid",
			},
		}
	}

	// Cek apakah nama alamat sudah pernah digunakan
	var countNamaSama int64
	if err := db.Model(&models.AlamatKurir{}).
		Where(&models.AlamatKurir{NamaAlamat: data.NamaAlamat}).
		Count(&countNamaSama).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseMasukanAlamatKurir{
				Message: "Gagal: Server sedang sibuk, coba lagi lain waktu",
			},
		}
	}

	if countNamaSama != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseMasukanAlamatKurir{
				Message: "Gagal: Kamu sudah memiliki alamat dengan nama itu, coba ganti nama lainnya",
			},
		}
	}

	// Simpan alamat baru
	if err := db.Create(&models.AlamatKurir{
		PanggilanAlamat: data.PanggilanAlamat,
		NomorTelephone:  data.NomorTelephone,
		NamaAlamat:      data.NamaAlamat,
		Kota:            data.Kota,
		KodeNegara:      data.KodeNegara,
		KodePos:         data.KodePos,
		Deskripsi:       data.Deskripsi,
		Longitude:       data.Longtitude,
		Latitude:        data.Latitude,
	}).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseMasukanAlamatKurir{
				Message: "Gagal: Tidak dapat menyimpan data alamat, coba lagi lain waktu",
			},
		}
	}

	// Berhasil
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_kurir.ResponseMasukanAlamatKurir{
			Message: "Berhasil",
		},
	}
}

func EditAlamatKurir(data PayloadEditAlamatKurir, db *gorm.DB) *response.ResponseForm {
	services := "EditAlamatKurir"

	// Validasi identitas kurir
	_, valid := data.IdentitasKurir.Validating(db)
	if !valid {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseEditAlamatKurir{
				Message: "Gagal: Data kurir tidak valid",
			},
		}
	}

	// Cek apakah alamat dengan ID dan kurir terkait benar-benar ada
	var countExist int64
	if err := db.Model(&models.AlamatKurir{}).
		Where(&models.AlamatKurir{
			ID:      data.IDAlamatKurir,
			IdKurir: data.IdentitasKurir.IdKurir,
		}).
		Count(&countExist).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseEditAlamatKurir{
				Message: "Gagal: Server sedang sibuk, coba lagi lain waktu",
			},
		}
	}

	if countExist == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseEditAlamatKurir{
				Message: "Gagal: Data alamat tidak ditemukan",
			},
		}
	}

	// Update data alamat
	if err := db.Model(&models.AlamatKurir{}).
		Where(&models.AlamatKurir{ID: data.IDAlamatKurir}).
		Updates(&models.AlamatKurir{
			PanggilanAlamat: data.PanggilanAlamat,
			NomorTelephone:  data.NomorTelephone,
			NamaAlamat:      data.NamaAlamat,
			Kota:            data.Kota,
			KodeNegara:      data.KodeNegara,
			KodePos:         data.KodePos,
			Deskripsi:       data.Deskripsi,
			Longitude:       data.Longtitude,
			Latitude:        data.Latitude,
		}).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseEditAlamatKurir{
				Message: "Gagal: Mengubah data alamat, coba lagi nanti",
			},
		}
	}

	// Berhasil
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_kurir.ResponseEditAlamatKurir{
			Message: "Berhasil",
		},
	}
}

func HapusAlamatKurir(data PayloadHapusAlamatKurir, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatKurir"

	// Validasi identitas kurir
	_, valid := data.IdentitasKurir.Validating(db)
	if !valid {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseHapusAlamatKurir{
				Message: "Gagal: Data kurir tidak valid",
			},
		}
	}

	// Cek apakah alamat milik kurir tersebut ada
	var countExist int64
	if err := db.Model(&models.AlamatKurir{}).
		Where(&models.AlamatKurir{
			ID:      data.IdAlamatKurir,
			IdKurir: data.IdentitasKurir.IdKurir,
		}).
		Count(&countExist).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseHapusAlamatKurir{
				Message: "Gagal: Tidak dapat memverifikasi data alamat, coba lagi nanti",
			},
		}
	}

	if countExist == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseHapusAlamatKurir{
				Message: "Gagal: Alamat itu tidak ada",
			},
		}
	}

	// Hapus alamat
	if err := db.Model(&models.AlamatKurir{}).
		Where(&models.AlamatKurir{
			ID:      data.IdAlamatKurir,
			IdKurir: data.IdentitasKurir.IdKurir,
		}).
		Delete(&models.AlamatKurir{}).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_kurir.ResponseHapusAlamatKurir{
				Message: "Gagal: Server sedang sibuk, coba lagi lain waktu",
			},
		}
	}

	// Berhasil
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_kurir.ResponseHapusAlamatKurir{
			Message: "Berhasil",
		},
	}
}
