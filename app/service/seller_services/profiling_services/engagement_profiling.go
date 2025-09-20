package seller_profiling_services

import (
	"context"
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_particular_profiling "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services/particular_profiling"
	seller_response_profiling "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services/response_profiling"
)

func UpdatePersonalSeller(ctx context.Context, db *gorm.DB, data PayloadUpdateProfilePersonalSeller) *response.ResponseForm {
	fmt.Println("Trace UpdatePersonalSeller:")
	services := "UpdatePersonalSeller"
	var hasil_update_nama *seller_particular_profiling.ResponseUbahNama
	var hasil_update_username *seller_particular_profiling.ResponseUbahUsername
	var hasil_update_gmail *seller_particular_profiling.ResponseUbahEmail

	if data.ID_Seller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
		}
	}

	if data.Username == "" {
		fmt.Println("❌ username: kosong (tidak diupdate)")
	} else {
		hasil_update_username = seller_particular_profiling.UbahUsernameSeller(data.ID_Seller, data.Username, db)
	}

	if data.Nama == "" {
		fmt.Println("❌ nama: kosong (tidak diupdate)")
	} else {
		hasil_update_nama = seller_particular_profiling.UbahNamaSeller(data.ID_Seller, data.Nama, db)
	}

	if data.Email == "" {
		fmt.Println("❌ email: kosong (tidak diupdate)")
	} else {
		hasil_update_gmail = seller_particular_profiling.UbahEmailSeller(ctx, data.ID_Seller, data.Email, db)
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

func UpdateInfoGeneralPublic(db *gorm.DB, data PayloadUpdateInfoGeneralSeller) *response.ResponseForm {
	services := "UpdatePersonalSeller"
	var hasil_update_punchline seller_particular_profiling.ResponseUbahPunchline
	var hasil_update_deskripsi seller_particular_profiling.ResponseUbahDeskripsi
	var hasil_update_jam_operasional seller_particular_profiling.ResponseUbahJamOperasional

	if data.ID_Seller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
		}
	}

	if data.Deskripsi != "" {
		hasil_update_deskripsi = *seller_particular_profiling.UbahDeskripsiSeller(data.ID_Seller, data.Username, data.Deskripsi, db)
	}

	if data.Punchline != "" {
		hasil_update_punchline = *seller_particular_profiling.UbahPunchlineSeller(data.ID_Seller, data.Username, data.Punchline, db)
	}

	if data.JamOperasional != "" {
		hasil_update_jam_operasional = *seller_particular_profiling.UbahJamOperasionalSeller(data.ID_Seller, data.Username, data.JamOperasional, db)
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
