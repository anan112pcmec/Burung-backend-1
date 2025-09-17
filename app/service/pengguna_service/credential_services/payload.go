package pengguna_credential_services

type PayloadPreUbahPasswordPengguna struct {
	IDPengguna      int64  `json:"id_user_ganti_password"`
	Username        string `json:"username_user_ganti_password"`
	PasswordSebelum string `json:"password_sebelum_user"`
	PasswordBaru    string `json:"password_baru_user"`
	FaktorKedua     string `json:"faktor_kedua_ganti_password_user"`
}

type PayloadValidateOTPPasswordPengguna struct {
	IDPengguna int64  `json:"id_user_validate"`
	OtpKey     string `json:"otp_key_ubah_password"`
}

type PayloadValidatePinPasswordPengguna struct {
	IDPengguna int64  `json:"id_user_validate_ganti_password"`
	Pin        string `json:"pin_key_ubah_password"`
}
