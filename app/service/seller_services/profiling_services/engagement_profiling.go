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
	var hasil_update_nama seller_particular_profiling.ResponseUbahNama
	var hasil_update_username seller_particular_profiling.ResponseUbahUsername
	var hasil_update_gmail seller_particular_profiling.ResponseUbahEmail

	if _, status := data.IdentitasSeller.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Username == "not" {
		log.Println("[INFO] Username kosong, tidak diupdate.")
	} else {
		hasil_update_username = seller_particular_profiling.UbahUsernameSeller(ctx, data.IdentitasSeller.IdSeller, data.Username, db)
	}

	if data.Nama == "not" {
		log.Println("[INFO] Nama kosong, tidak diupdate.")
	} else {
		hasil_update_nama = seller_particular_profiling.UbahNamaSeller(ctx, data.IdentitasSeller.IdSeller, data.Nama, db)
	}

	if data.Email == "not" {
		log.Println("[INFO] Email kosong, tidak diupdate.")
	} else {
		hasil_update_gmail = seller_particular_profiling.UbahEmailSeller(ctx, data.IdentitasSeller.IdSeller, data.Email, db)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: seller_response_profiling.ResponseUpdateProfileSeller{
			UpdateNama:     hasil_update_nama,
			UpdateUsername: hasil_update_username,
			UpdateGmail:    hasil_update_gmail,
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Update Info General Public
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UpdateInfoGeneralPublic(ctx context.Context, db *gorm.DB, data PayloadUpdateInfoGeneralSeller) *response.ResponseForm {
	services := "UpdatePersonalSeller"
	var hasil_update_punchline seller_particular_profiling.ResponseUbahPunchline
	var hasil_update_deskripsi seller_particular_profiling.ResponseUbahDeskripsi
	var hasil_update_jam_operasional seller_particular_profiling.ResponseUbahJamOperasional
	var hasil_update_dedication seller_particular_profiling.ResponseUbahDedication

	seller, status := data.IdentitasSeller.Validating(ctx, db)
	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Deskripsi != "not" && data.Deskripsi != "" && data.Deskripsi != seller.Deskripsi {
		hasil_update_deskripsi = seller_particular_profiling.UbahDeskripsiSeller(ctx, data.IdentitasSeller.IdSeller, data.Deskripsi, db)
	} else {
		log.Println("[INFO] Deskripsi kosong atau tidak berubah, tidak diupdate.")
	}

	if data.Punchline != "not" && data.Punchline != "" && data.Punchline != seller.Punchline {
		hasil_update_punchline = seller_particular_profiling.UbahPunchlineSeller(ctx, data.IdentitasSeller.IdSeller, data.Punchline, db)
	} else {
		log.Println("[INFO] Punchline kosong atau tidak berubah, tidak diupdate.")
	}

	if data.JamOperasional != "not" && data.JamOperasional != "" && data.JamOperasional != seller.JamOperasional {
		hasil_update_jam_operasional = seller_particular_profiling.UbahJamOperasionalSeller(ctx, data.IdentitasSeller.IdSeller, data.JamOperasional, db)
	} else {
		log.Println("[INFO] Jam operasional kosong atau tidak berubah, tidak diupdate.")
	}

	if data.Dedication != "not" && data.Dedication != "" && data.Dedication != seller.SellerDedication {
		hasil_update_dedication = seller_particular_profiling.UbahSellerDedication(ctx, data.IdentitasSeller.IdSeller, data.Dedication, db)
	} else {
		log.Println("[INFO] Dedication seller kosong atau tidak berubah, tidak diupdate.")
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: seller_response_profiling.ResponseUpdateInfoGeneralSeller{
			UpdatePunchline:      hasil_update_punchline,
			UpdateDeskripsi:      hasil_update_deskripsi,
			UpdateJamOperasional: hasil_update_jam_operasional,
			UpdateDedication:     hasil_update_dedication,
		},
	}
}
