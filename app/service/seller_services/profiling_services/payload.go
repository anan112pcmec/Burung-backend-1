package seller_profiling_services

type KredensialSeller struct {
	ID_Seller int32  `json:"id_seller"`
	Username  string `json:"username_seller"`
}

type PayloadUpdateProfilePersonalSeller struct {
	KredensialSeller
	Username string `json:"update_username_seller"`
	Nama     string `json:"update_nama_seller"`
	Email    string `json:"update_email_seller"`
}

type PayloadUpdateInfoGeneralSeller struct {
	KredensialSeller
	JamOperasional string `json:"update_jam_operasional_seller"`
	Punchline      string `json:"update_punchline_seller"`
	Deskripsi      string `json:"update_deskripsi_seller"`
}
