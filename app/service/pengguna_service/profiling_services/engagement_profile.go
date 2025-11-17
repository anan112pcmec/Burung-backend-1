package pengguna_profiling_services

import (
	"context"
	"net/http"
	"sync"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	particular_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/particular_profiling"
	response_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/response_profiling"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Ubah Personal Profiling Pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahPersonalProfilingPengguna(ctx context.Context, data PayloadPersonalProfilingPengguna, db *gorm.DB) *response.ResponseForm {
	var wg sync.WaitGroup

	services := "UbahPersonalProfilingPengguna"
	var hasil_update_gmail particular_profiling_pengguna.ResponseUbahEmail
	var hasil_update_username particular_profiling_pengguna.ResponseUbahUsername
	var hasil_update_nama particular_profiling_pengguna.ResponseUbahNama

	seller, status := data.IdentitasPengguna.Validating(ctx, db)
	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.EmailUpdate != "" && data.EmailUpdate != seller.Email && data.EmailUpdate != "not" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasil_update_gmail = particular_profiling_pengguna.UbahEmailPengguna(ctx, data.IdentitasPengguna.ID, data.EmailUpdate, db)
		}()
	}

	if data.UsernameUpdate != "" && data.UsernameUpdate != seller.Username && data.UsernameUpdate != "not" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasil_update_username = particular_profiling_pengguna.UbahUsernamePengguna(ctx, db, data.IdentitasPengguna.ID, data.UsernameUpdate)
		}()
	}

	if data.NamaUpdate != "" && data.NamaUpdate != seller.Nama && data.NamaUpdate != "not" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasil_update_nama = particular_profiling_pengguna.UbahNamaPengguna(ctx, data.IdentitasPengguna.ID, data.NamaUpdate, db)
		}()
	}

	wg.Wait()

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
