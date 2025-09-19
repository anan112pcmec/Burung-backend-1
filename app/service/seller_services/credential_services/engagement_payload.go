package seller_credential_services

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

///////////////////////////////////////////////////////////////////////
// PASSWORD SELLER
// ////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////
// NOREK SELLER
// ////////////////////////////////////////////////////////////////////

type PayloadTambahkanNorekSeller struct {
	Data models.RekeningSeller `json:"data_rekening_seller"`
}

type PayloadHapusNorekSeller struct {
	IDSeller        int32  `json:"id_seller_hapus_rekening"`
	NamaBank        string `json:"nama_bank_hapus_rekening"`
	NomorRekening   string `json:"nomor_rekening_hapus_rekening"`
	PemilikRekening string `json:"pemilik_rekening_hapus_rekening"`
}
