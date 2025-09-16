package seller_profiling_services

type IdSeller struct {
	ID_Seller int32 `json:"id_seller"`
}

type PayloadUpdateProfilePersonalSeller struct {
	IdSeller
	Username string `json:"update_username_seller"`
	Nama     string `json:"update_nama_seller"`
	Email    string `json:"update_email_seller"`
}

type PayloadUpdateProfileGeneralSeller struct {
	IdSeller
	SellerDedication string `json:"update_seller_dedication_seller"`
	JamOperasional   string `json:"udpate_jam_operasional_seller"`
	Punchline        string `json:"update_punchline_seller"`
	Deskripsi        string `json:"update_deskripsi_seller"`
}
