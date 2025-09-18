package pengguna_credential_services

// ////////////////////////////////////////////////////////////////////////////////////////////////////////
// PAYLOAD ENGAGEMENT PASSWORD
// ////////////////////////////////////////////////////////////////////////////////////////////////////////

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
	IDPengguna int64  `json:"id_user_validate"`
	Pin        string `json:"pin_key_ubah_password"`
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////
// PAYLOAD ENGAGEMENT PIN
// ////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadMembuatPinPengguna struct {
	IDPengguna int64  `json:"id_user_membuat_pin"`
	Username   string `json:"username_user_membuat_pin"`
	Password   string `json:"password_user_membuat_pin"`
	Pin        string `json:"pin_user_membuat_pin"`
}

type PayloadUpdatePinPengguna struct {
	IDPengguna int64  `json:"id_user_update_pin"`
	Username   string `json:"username_user_update_pin"`
	PinBaru    string `json:"pin_baru_update_pin"`
	PinLama    string `json:"pin_lama_update_pin"`
}
