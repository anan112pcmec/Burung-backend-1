package response_profiling_pengguna

import particular_profiling_pengguna "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services/particular_profiling"

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct Personal Profiling Pengguna
// Melakukan Merge pada response particular profiling response
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ResponsePersonalProfilingPengguna struct {
	UpdateUsername particular_profiling_pengguna.ResponseUbahUsername `json:"response_update_username_user"`
	UpdateNama     particular_profiling_pengguna.ResponseUbahNama     `json:"response_update_nama_user"`
	UpdateGmail    particular_profiling_pengguna.ResponseUbahEmail    `json:"response_update_gmail_user"`
}
