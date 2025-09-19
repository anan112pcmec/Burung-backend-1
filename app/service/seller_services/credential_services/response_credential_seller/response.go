package response_credential_seller

///////////////////////////////////////////////////////////////////////
// PASSWORD SELLER
// ////////////////////////////////////////////////////////////////////

type ResponsePreUbahPasswordSeller struct {
	Message string `json:"pesan_preupdate_password_seller"`
}

type ResponseValidateUbahPasswordSeller struct {
	Message string `json:"pesan_validate_password_seller"`
}

///////////////////////////////////////////////////////////////////////
// REKENING SELLER
// ////////////////////////////////////////////////////////////////////

type ResponseTambahRekeningSeller struct {
	Message string `json:"pesan_tambah_rekening_seller"`
}

type ResponseHapusRekeningSeller struct {
	Message string `json:"pesan_hapus_rekening_seller"`
}
