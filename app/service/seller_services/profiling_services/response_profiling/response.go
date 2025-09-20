package seller_response_profiling

import seller_particular_profiling "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services/particular_profiling"

type ResponseUpdateProfileSeller struct {
	UpdateUsername seller_particular_profiling.ResponseUbahUsername `json:"response_update_username_seller"`
	UpdateNama     seller_particular_profiling.ResponseUbahNama     `json:"response_update_nama_seller"`
	UpdateGmail    seller_particular_profiling.ResponseUbahEmail    `json:"response_update_gmail_seller"`
}

type ResponseUpdateInfoGeneralSeller struct {
	UpdatePunchline      seller_particular_profiling.ResponseUbahPunchline      `json:"response_update_punchline_seller"`
	UpdateDeskripsi      seller_particular_profiling.ResponseUbahDeskripsi      `json:"response_update_deskripsi_seller"`
	UpdateJamOperasional seller_particular_profiling.ResponseUbahJamOperasional `json:"response_update_jam_operasional_seller"`
}
