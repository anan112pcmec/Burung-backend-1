package seller_credential_services

type PayloadPreUbahPasswordSeller struct {
	IDSeller     int32  `json:"id_seller_ganti_password"`
	Username     string `json:"username_seller_ganti_password"`
	PasswordLama string `json:"password_lama_ganti_password"`
	PasswordBaru string `json:"password_baru_ganti_password"`
}

type PayloadValidateUbahPasswordSellerOTP struct {
	IDSeller             int32  `json:"id_seller_validate"`
	OtpKeyValidateSeller string `json:"otp_key_validate_ubah_password_seller"`
}
