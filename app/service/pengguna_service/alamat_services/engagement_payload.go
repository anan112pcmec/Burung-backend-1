package pengguna_alamat_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"
)

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Masukan Alamat Pengguna
// ///////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadMasukanAlamatPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	DataAlamat        models.AlamatPengguna              `json:"data_alamat_pengguna"`
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Hapus Alamat Pengguna
// ///////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadHapusAlamatPengguna struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	IdAlamat          int64                              `json:"id_alamat_hapus_alamat"`
}
