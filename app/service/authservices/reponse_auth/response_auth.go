package response_auth

type LoginResponse struct {
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// /////////////////////////////////////////

// USER RESPONSE

// /////////////////////////////////////////

type LoginUserResp struct {
	Status  string `json:"status_login_user"`
	Message string `json:"pesan_login_user"`
	ID      int64  `json:"id_user"`
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
	ID      int32  `json:"id_seller"`
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
