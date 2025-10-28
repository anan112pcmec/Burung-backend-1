package seller_profiling_services

import (
	"context"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_particular_profiling "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services/particular_profiling"
	seller_response_profiling "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services/response_profiling"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Update Personal Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UpdatePersonalSeller(ctx context.Context, db *gorm.DB, data PayloadUpdateProfilePersonalSeller) *response.ResponseForm {
	services := "UpdatePersonalSeller"
	var hasil_update_nama *seller_particular_profiling.ResponseUbahNama
	var hasil_update_username *seller_particular_profiling.ResponseUbahUsername
	var hasil_update_gmail *seller_particular_profiling.ResponseUbahEmail

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Username == "" {
		log.Println("[INFO] Username kosong, tidak diupdate.")
	} else {
		hasil_update_username = seller_particular_profiling.UbahUsernameSeller(data.IdentitasSeller.IdSeller, data.Username, db)
	}

	if data.Nama == "" {
		log.Println("[INFO] Nama kosong, tidak diupdate.")
	} else {
		hasil_update_nama = seller_particular_profiling.UbahNamaSeller(data.IdentitasSeller.IdSeller, data.Nama, db)
	}

	if data.Email == "" {
		log.Println("[INFO] Email kosong, tidak diupdate.")
	} else {
		hasil_update_gmail = seller_particular_profiling.UbahEmailSeller(ctx, data.IdentitasSeller.IdSeller, data.Email, db)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: seller_response_profiling.ResponseUpdateProfileSeller{
			UpdateNama:     *hasil_update_nama,
			UpdateUsername: *hasil_update_username,
			UpdateGmail:    *hasil_update_gmail,
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Update Info General Public
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UpdateInfoGeneralPublic(db *gorm.DB, data PayloadUpdateInfoGeneralSeller) *response.ResponseForm {
	services := "UpdatePersonalSeller"
	var hasil_update_punchline seller_particular_profiling.ResponseUbahPunchline
	var hasil_update_deskripsi seller_particular_profiling.ResponseUbahDeskripsi
	var hasil_update_jam_operasional seller_particular_profiling.ResponseUbahJamOperasional

	if _, status := data.IdentitasSeller.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Deskripsi != "" {
		hasil_update_deskripsi = *seller_particular_profiling.UbahDeskripsiSeller(data.IdentitasSeller.IdSeller, data.IdentitasSeller.Username, data.Deskripsi, db)
	} else {
		log.Println("[INFO] Deskripsi kosong, tidak diupdate.")
	}

	if data.Punchline != "" {
		hasil_update_punchline = *seller_particular_profiling.UbahPunchlineSeller(data.IdentitasSeller.IdSeller, data.IdentitasSeller.Username, data.Punchline, db)
	} else {
		log.Println("[INFO] Punchline kosong, tidak diupdate.")
	}

	if data.JamOperasional != "" {
		hasil_update_jam_operasional = *seller_particular_profiling.UbahJamOperasionalSeller(data.IdentitasSeller.IdSeller, data.IdentitasSeller.Username, data.JamOperasional, db)
	} else {
		log.Println("[INFO] Jam operasional kosong, tidak diupdate.")
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: seller_response_profiling.ResponseUpdateInfoGeneralSeller{
			UpdatePunchline:      hasil_update_punchline,
			UpdateDeskripsi:      hasil_update_deskripsi,
			UpdateJamOperasional: hasil_update_jam_operasional,
		},
	}
}
