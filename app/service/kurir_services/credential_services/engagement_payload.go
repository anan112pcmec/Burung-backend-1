package kurir_credential_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type PayloadPreUbahPassword struct {
	DataIdentitas identity_kurir.IdentitasKurir `json:"data_identitas_kurir"`
	PasswordLama  string                        `json:"password_lama_kurir"`
	PasswordBaru  string                        `json:"password_baru_kurir"`
}

type PayloadValidateUbahPassword struct {
	DataIdentitas identity_kurir.IdentitasKurir `json:"data_identitas_kurir"`
	OtpKey        string                        `json:"otp_key_ubah_password"`
}
