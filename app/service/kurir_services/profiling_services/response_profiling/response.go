package response_profiling_kurir

import particular_profiling_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/profiling_services/particular_profiling"

type ResponseProfilingGeneralKurir struct {
	UpdateNama     particular_profiling_kurir.ResponseUbahNama     `json:"response_ubah_nama_kurir"`
	UpdateUsername particular_profiling_kurir.ResponseUbahUsername `json:"response_ubah_username_kurir"`
	UpdateEmail    particular_profiling_kurir.ResponseUbahGmail    `json:"response_ubah_email_kurir"`
}
