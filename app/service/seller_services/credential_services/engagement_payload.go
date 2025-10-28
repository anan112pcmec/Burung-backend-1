package seller_credential_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

///////////////////////////////////////////////////////////////////////
// Payload Struct Pre Ubah Password Seller
// ////////////////////////////////////////////////////////////////////

type PayloadPreUbahPasswordSeller struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	PasswordLama    string                         `json:"password_lama_ganti_password"`
	PasswordBaru    string                         `json:"password_baru_ganti_password"`
}

///////////////////////////////////////////////////////////////////////
// Payload Struct Validate Ubah Password Seller
// ////////////////////////////////////////////////////////////////////

type PayloadValidateUbahPasswordSellerOTP struct {
	IDSeller             int32  `json:"id_seller_validate"`
	OtpKeyValidateSeller string `json:"otp_key_validate_ubah_password_seller"`
}

///////////////////////////////////////////////////////////////////////
// Payload Struct Tambahkan Nomor Rekening Seller
// ////////////////////////////////////////////////////////////////////

type PayloadTambahkanNorekSeller struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	Data            models.RekeningSeller          `json:"data_rekening_seller"`
}

///////////////////////////////////////////////////////////////////////
// Payload Hapus Nomor Rekening Seller
// ////////////////////////////////////////////////////////////////////

type PayloadHapusNorekSeller struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	NamaBank        string                         `json:"nama_bank_hapus_rekening"`
	NomorRekening   string                         `json:"nomor_rekening_hapus_rekening"`
	PemilikRekening string                         `json:"pemilik_rekening_hapus_rekening"`
}
