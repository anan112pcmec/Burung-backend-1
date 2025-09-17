package pengguna_profiling_services

type PayloadPersonalProfilingPengguna struct {
	IDPengguna int64  `json:"id_user_update_profile"`
	Username   string `json:"update_username_user"`
	Nama       string `json:"update_nama_user"`
	Email      string `json:"update_email_user"`
}
