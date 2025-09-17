package response_credential_pengguna

type ResponsePreUbahPassword struct {
	Message string `json:"pesan_preubah_password_user"`
}

type ResponseValidatePassword struct {
	Message string `json:"pesan_merubah_password_user"`
}
