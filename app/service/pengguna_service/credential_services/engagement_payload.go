package pengguna_credential_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////
// PAYLOAD ENGAGEMENT PASSWORD
// ////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadPreUbahPasswordPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	PasswordSebelum   string                             `json:"password_sebelum_user"`
	PasswordBaru      string                             `json:"password_baru_user"`
	FaktorKedua       string                             `json:"faktor_kedua_ganti_password_user"`
}

type PayloadValidateOTPPasswordPengguna struct {
	IDPengguna int64  `json:"id_pengguna"`
	OtpKey     string `json:"otp_key_ubah_password"`
}

type PayloadValidatePinPasswordPengguna struct {
	IDPengguna int64  `json:"id_pengguna"`
	Pin        string `json:"pin_key_ubah_password"`
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////
// PAYLOAD ENGAGEMENT PIN
// ////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadMembuatPinPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	Password          string                             `json:"password_user_membuat_pin"`
	Pin               string                             `json:"pin_user_membuat_pin"`
}

type PayloadUpdatePinPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	PinBaru           string                             `json:"pin_baru_update_pin"`
	PinLama           string                             `json:"pin_lama_update_pin"`
}
