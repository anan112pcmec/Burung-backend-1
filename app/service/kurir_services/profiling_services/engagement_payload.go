package kurir_profiling_service

type KredensialKurir struct {
	IDkurir       int64  `json:"id_kurir"`
	UsernameKurir string `json:"username_kurir"`
}

type PayloadPersonalProfilingKurir struct {
	DataKredensial KredensialKurir `json:"data_kredensial_kurir"`
	Username       string          `json:"ubah_username_kurir"`
	Nama           string          `json:"ubah_nama_kurir"`
	Email          string          `json:"ubah_email_kurir"`
}
