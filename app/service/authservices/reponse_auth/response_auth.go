package response_auth

type LoginResponse struct {
	ID       int64  `json:"id"`
	Nama     string `json:"nama"`
	Username string `json:"username"`
	JWT      string `json:"jwt_key"`
}

// /////////////////////////////////////////
// USER RESPONSE
// /////////////////////////////////////////

type LoginUserResp struct {
	Status  string `json:"status_login_user"`
	Message string `json:"pesan_login_user"`
	LoginResponse
}

type PreRegistrationUserResp struct {
	Status  string `json:"status_pre_registrasi_user"`
	Message string `json:"pesan_pre_registrasi_user"`
}

type ValidateUserResp struct {
	Status  string `json:"status_validasi_user"`
	Message string `json:"pesan_validasi_user"`
}

// /////////////////////////////////////////
// SELLER RESPONSE
// /////////////////////////////////////////

type LoginSellerResp struct {
	Status  string `json:"status_login_seller"`
	Message string `json:"pesan_login_seller"`
	LoginResponse
}

type PreRegistrationSellerResp struct {
	Status  string `json:"status_pre_registrasi_seller"`
	Message string `json:"pesan_pre_registrasi_seller"`
}

type ValidateSellerResp struct {
	Status  string `json:"status_validasi_seller"`
	Message string `json:"pesan_validasi_seller"`
}

// /////////////////////////////////////////
// KURIR RESPONSE
// /////////////////////////////////////////

type LoginKurirResp struct {
	Status  string `json:"status_login_kurir"`
	Message string `json:"pesan_login_kurir"`
	LoginResponse
}

type PreRegistrationKurirResp struct {
	Status  string `json:"status_pre_registrasi_kurir"`
	Message string `json:"pesan_pre_registrasi_kurir"`
}

type ValidateKurirResp struct {
	Status  string `json:"status_validasi_kurir"`
	Message string `json:"pesan_validasi_kurir"`
}
