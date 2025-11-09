package pengguna_profiling_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	particular_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/particular_profiling"
	response_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/response_profiling"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Ubah Personal Profiling Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahPersonalProfilingPengguna(ctx context.Context, data PayloadPersonalProfilingPengguna, db *gorm.DB) *response.ResponseForm {
	services := "UbahPersonalProfilingPengguna"
	var hasil_update_gmail particular_profiling_pengguna.ResponseUbahEmail
	var hasil_update_username particular_profiling_pengguna.ResponseUbahUsername
	var hasil_update_nama particular_profiling_pengguna.ResponseUbahNama

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	hasil_update_gmail = *particular_profiling_pengguna.UbahEmailPengguna(ctx, data.IdentitasPengguna.ID, data.EmailUpdate, db)

	hasil_update_username = *particular_profiling_pengguna.UbahUsernamePengguna(ctx, db, data.IdentitasPengguna.ID, data.UsernameUpdate)

	hasil_update_nama = *particular_profiling_pengguna.UbahNamaPengguna(ctx, data.IdentitasPengguna.ID, data.NamaUpdate, db)

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
