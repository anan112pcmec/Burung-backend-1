package pengguna_profiling_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Personal Profiling pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadPersonalProfilingPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	UsernameUpdate    string                             `json:"update_username_user"`
	NamaUpdate        string                             `json:"update_nama_user"`
	EmailUpdate       string                             `json:"update_email_user"`
}
