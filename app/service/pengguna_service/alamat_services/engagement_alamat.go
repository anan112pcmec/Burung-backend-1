package pengguna_alamat_services

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/alamat_services/response_alamat_service_pengguna"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Masukan Alamat Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanAlamatPengguna(data PayloadMasukanAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "MasukanAlamatPengguna"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
				Messages: "Gagal Identitas Kamu Tidak Selaras Dengan Target.",
			},
		}
	}

	var count int64
	if err := db.Model(&models.AlamatPengguna{}).
		Where(models.AlamatPengguna{IDPengguna: data.DataAlamat.IDPengguna}).
		Count(&count).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Coba Lagi Nanti Server Sedang Sibuk",
		}
	}

	if count >= 5 {
		return &response.ResponseForm{
			Status:   http.StatusForbidden,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
				Messages: "Batas Maksimum Penyimpanan Alamat Tercapai (Maksimal 5 Alamat).",
			},
		}
	}

	var existing models.AlamatPengguna
	errcheck := db.Where(models.AlamatPengguna{
		IDPengguna:      data.IdentitasPengguna.ID,
		PanggilanAlamat: data.DataAlamat.PanggilanAlamat,
	}).First(&existing).Error

	if errors.Is(errcheck, gorm.ErrRecordNotFound) {
		if err := db.Create(&models.AlamatPengguna{
			IDPengguna:      data.IdentitasPengguna.ID,
			PanggilanAlamat: data.DataAlamat.PanggilanAlamat,
			NamaAlamat:      data.DataAlamat.NamaAlamat,
			Deskripsi:       data.DataAlamat.Deskripsi,
			NomorTelephone:  data.DataAlamat.NomorTelephone,
			Kota:            data.DataAlamat.Kota,
			KodePos:         data.DataAlamat.KodePos,
			KodeNegara:      data.DataAlamat.KodeNegara,
			Longitude:       data.DataAlamat.Longitude,
			Latitude:        data.DataAlamat.Latitude,
		}).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
					Messages: "Terjadi Kesalahan Pada Server. Silakan Coba Beberapa Saat Lagi.",
				},
			}
		}
	} else {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
				Messages: "Alamat Dengan Panggilan Tersebut Sudah Ada. Silakan Gunakan Panggilan Lain.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseMembuatAlamat{
			Messages: "Alamat Berhasil Ditambahkan.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Hapus Alamat Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusAlamatPengguna(data PayloadHapusAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatPengguna"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamat{
				Messages: "Gagal Menghapus Alamat, Identitas Mu Tidak Sesuai.",
			},
		}
	}

	if err_hapus := db.Where(models.AlamatPengguna{
		IDPengguna:      data.IdentitasPengguna.ID,
		PanggilanAlamat: data.PanggilanAlamat,
	}).Delete(&models.AlamatPengguna{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamat{
				Messages: "Terjadi Kesalahan Pada Server Saat Menghapus Alamat. Silakan Coba Beberapa Saat Lagi.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseHapusAlamat{
			Messages: "Alamat Berhasil Dihapus.",
		},
	}
}
