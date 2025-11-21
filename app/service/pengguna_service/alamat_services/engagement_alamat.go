package pengguna_alamat_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/alamat_services/response_alamat_service_pengguna"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Masukan Alamat Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func MasukanAlamatPengguna(ctx context.Context, data PayloadMasukanAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "MasukanAlamatPengguna"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamatPengguna{
				Messages: "Gagal Identitas Kamu Tidak Selaras Dengan Target.",
			},
		}
	}

	var id_data_alamats []int64
	if err := db.WithContext(ctx).Select("id").Model(&models.AlamatPengguna{}).
		Where(models.AlamatPengguna{IDPengguna: data.IdentitasPengguna.ID}).
		Limit(5).Scan(&id_data_alamats).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Coba Lagi Nanti Server Sedang Sibuk",
		}
	}

	if len(id_data_alamats) == 5 {
		return &response.ResponseForm{
			Status:   http.StatusForbidden,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamatPengguna{
				Messages: "Batas Maksimum Penyimpanan Alamat Tercapai (Maksimal 5 Alamat).",
			},
		}
	}

	var id_data_alamat int64 = 0
	if err := db.WithContext(ctx).Model(&models.AlamatPengguna{}).Select("id").Where(&models.AlamatPengguna{
		IDPengguna:      data.IdentitasPengguna.ID,
		PanggilanAlamat: data.PanggilanAlamat,
	}).Limit(1).Scan(&id_data_alamat).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_alamat != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal kamu sudah memiliki alamat dengan panggilan yang sama",
		}
	}

	helper.SanitasiKoordinat(&data.Latitude, &data.Longitude)

	if err := db.WithContext(ctx).Create(&models.AlamatPengguna{
		IDPengguna:      data.IdentitasPengguna.ID,
		PanggilanAlamat: data.PanggilanAlamat,
		NamaAlamat:      data.NamaAlamat,
		Deskripsi:       data.Deskripsi,
		NomorTelephone:  data.NomorTelephone,
		Provinsi:        data.Provinsi,
		Kota:            data.Kota,
		KodePos:         data.KodePos,
		KodeNegara:      data.KodeNegara,
		Longitude:       data.Longitude,
		Latitude:        data.Latitude,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseMembuatAlamatPengguna{
			Messages: "Alamat Berhasil Ditambahkan.",
		},
	}
}

func EditAlamatPengguna(ctx context.Context, data PayloadEditAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "EditAlamatPengguna"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseMembuatAlamatPengguna{
				Messages: "Gagal Identitas Kamu Tidak Selaras Dengan Target.",
			},
		}
	}

	var id_alamat_pengguna int64 = 0
	if err := db.WithContext(ctx).Model(&models.AlamatPengguna{}).Select("id").Where(&models.AlamatPengguna{
		ID:         data.IdAlamatPengguna,
		IDPengguna: data.IdentitasPengguna.ID,
	}).Limit(1).Scan(&id_alamat_pengguna).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseEditAlamatPengguna{
				Message: "Gagal Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_alamat_pengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseEditAlamatPengguna{
				Message: "Gagal Data Alamat Tidak Valid",
			},
		}
	}

	var idDataTransaksi int64 = 0

	if err := db.WithContext(ctx).
		Model(&models.Transaksi{}).
		Select("id").
		Where("id_alamat_pengguna = ? AND status != ?", data.IdAlamatPengguna, "Selesai").
		Limit(1).
		Scan(&idDataTransaksi).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal, server sedang sibuk. Coba lagi lain waktu",
		}
	}

	// Jika ada transaksi yang menggunakan alamat ini
	if idDataTransaksi != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal, alamat sedang digunakan sebagai acuan transaksi",
		}
	}

	helper.SanitasiKoordinat(&data.Latitude, &data.Longitude)

	if err := db.WithContext(ctx).Model(&models.AlamatPengguna{}).Where(&models.AlamatPengguna{
		ID: data.IdAlamatPengguna,
	}).Updates(&models.AlamatPengguna{
		PanggilanAlamat: data.PanggilanAlamat,
		NamaAlamat:      data.NamaAlamat,
		Deskripsi:       data.Deskripsi,
		NomorTelephone:  data.NomorTelephone,
		Provinsi:        data.Provinsi,
		Kota:            data.Kota,
		KodePos:         data.KodePos,
		KodeNegara:      data.KodeNegara,
		Longitude:       data.Longitude,
		Latitude:        data.Latitude,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseEditAlamatPengguna{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseEditAlamatPengguna{
			Message: "Berhasil",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Hapus Alamat Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusAlamatPengguna(ctx context.Context, data PayloadHapusAlamatPengguna, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatPengguna"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusOK,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamatPengguna{
				Message: "Gagal Menghapus Alamat, Identitas Mu Tidak Sesuai.",
			},
		}
	}

	var id_alamat_pengguna int64 = 0
	if err := db.WithContext(ctx).Model(&models.AlamatPengguna{}).Select("id").Where(&models.AlamatPengguna{
		ID:         data.IdAlamatPengguna,
		IDPengguna: data.IdentitasPengguna.ID,
	}).Limit(1).Scan(&id_alamat_pengguna).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamatPengguna{
				Message: "Gagal Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_alamat_pengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamatPengguna{
				Message: "Gagal Data Alamat Tidak Valid",
			},
		}
	}

	var idDataTransaksi int64 = 0

	if err := db.WithContext(ctx).
		Model(&models.Transaksi{}).
		Select("id").
		Where("id_alamat_pengguna = ? AND status != ?", data.IdAlamatPengguna, "Selesai").
		Limit(1).
		Scan(&idDataTransaksi).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal, server sedang sibuk. Coba lagi lain waktu",
		}
	}

	// Jika ada transaksi yang menggunakan alamat ini
	if idDataTransaksi != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal, alamat sedang digunakan sebagai acuan transaksi",
		}
	}

	if err_hapus := db.WithContext(ctx).Where(models.AlamatPengguna{
		ID: data.IdAlamatPengguna,
	}).Delete(&models.AlamatPengguna{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_service_pengguna.ResponseHapusAlamatPengguna{
				Message: "Terjadi Kesalahan Pada Server Saat Menghapus Alamat. Silakan Coba Beberapa Saat Lagi.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_service_pengguna.ResponseHapusAlamatPengguna{
			Message: "Alamat Berhasil Dihapus.",
		},
	}
}
