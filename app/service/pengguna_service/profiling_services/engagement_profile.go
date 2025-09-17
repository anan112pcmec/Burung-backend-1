package pengguna_profiling_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	particular_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/particular_profiling"
	response_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/response_profiling"
)

func UbahPersonalProfilingPengguna(ctx context.Context, data PayloadPersonalProfilingPengguna, db *gorm.DB) *response.ResponseForm {
	services := "UbahPersonalProfilingPengguna"
	var hasil_update_gmail particular_profiling_pengguna.ResponseUbahEmail
	var hasil_update_username particular_profiling_pengguna.ResponseUbahUsername
	var hasil_update_nama particular_profiling_pengguna.ResponseUbahNama

	if data.IDPengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Email != "" {
		hasil_update_gmail = *particular_profiling_pengguna.UbahEmailSeller(ctx, data.IDPengguna, data.Email, db)
	}

	if data.Username != "" {
		hasil_update_username = *particular_profiling_pengguna.UbahUsernamePengguna(db, data.IDPengguna, data.Username)
	}

	if data.Nama != "" {
		hasil_update_nama = *particular_profiling_pengguna.UbahNamaPengguna(data.IDPengguna, data.Nama, db)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_profiling_pengguna.ResponsePersonalProfilingPengguna{
			UpdateNama:     hasil_update_nama,
			UpdateUsername: hasil_update_username,
			UpdateGmail:    hasil_update_gmail,
		},
	}
}
