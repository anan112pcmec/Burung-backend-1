package kurir_rekening_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/rekening_services/response_rekening_services_kurir"
)

func MasukanRekeningKurir(ctx context.Context, data PayloadMasukanRekeningKurir, db *gorm.DB) *response.ResponseForm {
	services := "MasukanRekeningKurir"

	_, validasi := data.IdentitasKurir.Validating(ctx, db)
	if !validasi {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Gagal Data Kurir Tidak Valid",
			},
		}
	}

	var id_alamat int64
	if err := db.WithContext(ctx).
		Model(&models.RekeningKurir{}).
		Select("id").
		Where(&models.RekeningKurir{IdKurir: data.IdentitasKurir.IdKurir}).
		Limit(1).
		Scan(&id_alamat).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Gagal, server sedang sibuk coba lagi nanti",
			},
		}
	}

	if id_alamat != 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Maksimal hanya memasukan 1 rekening",
			},
		}
	}

	if err := db.WithContext(ctx).Create(&models.RekeningKurir{
		IdKurir:         data.IdentitasKurir.IdKurir,
		NamaBank:        data.NamaBank,
		NomorRekening:   data.NomorRekening,
		PemilikRekening: data.PemilikRekening,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Gagal, server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
			Message: "Berhasil",
		},
	}
}

func EditRekeningKurir(ctx context.Context, data PayloadEditRekeningKurir, db *gorm.DB) *response.ResponseForm {
	services := "EditRekeningKurir"

	_, validasi := data.IdentitasKurir.Validating(ctx, db)
	if !validasi {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseEditRekeningKurir{
				Message: "Gagal menemukan data kurir",
			},
		}
	}

	var id_alamat int64
	if err := db.WithContext(ctx).
		Model(&models.RekeningKurir{}).
		Select("id").
		Where(&models.RekeningKurir{ID: data.IdRekening, IdKurir: data.IdentitasKurir.IdKurir}).
		Limit(1).
		Scan(&id_alamat).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Gagal, server sedang sibuk coba lagi nanti",
			},
		}
	}

	if id_alamat == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Maksimal hanya memasukan 1 rekening",
			},
		}
	}

	if err := db.WithContext(ctx).Model(&models.RekeningKurir{}).Where(&models.RekeningKurir{
		ID:      data.IdRekening,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Updates(&models.RekeningKurir{
		NamaBank:        data.NamaBank,
		NomorRekening:   data.NomorRekening,
		PemilikRekening: data.PemilikRekening,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseEditRekeningKurir{
				Message: "Gagal, server sedang sibuk, coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_rekening_services_kurir.ResponseEditRekeningKurir{
			Message: "Berhasil",
		},
	}
}

func HapusRekeningKurir(ctx context.Context, data PayloadHapusRekeningKurir, db *gorm.DB) *response.ResponseForm {
	services := "HapusRekeningKurir"

	_, validasi := data.IdentitasKurir.Validating(ctx, db)
	if !validasi {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseHapusRekeningKurir{
				Message: "Gagal menemukan data kurir",
			},
		}
	}

	var id_alamat int64
	if err := db.WithContext(ctx).
		Model(&models.RekeningKurir{}).
		Select("id").
		Where(&models.RekeningKurir{ID: data.IdRekening, IdKurir: data.IdentitasKurir.IdKurir}).
		Limit(1).
		Scan(&id_alamat).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Gagal, server sedang sibuk coba lagi nanti",
			},
		}
	}

	if id_alamat == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseMasukanRekeningKurir{
				Message: "Data Rekening Tidak Ditemukan",
			},
		}
	}

	if err := db.Model(&models.RekeningKurir{}).Where(&models.RekeningKurir{
		ID:      data.IdRekening,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Delete(&models.RekeningKurir{}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_rekening_services_kurir.ResponseHapusRekeningKurir{
				Message: "Gagal, server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_rekening_services_kurir.ResponseHapusRekeningKurir{
			Message: "Berhasil",
		},
	}
}
