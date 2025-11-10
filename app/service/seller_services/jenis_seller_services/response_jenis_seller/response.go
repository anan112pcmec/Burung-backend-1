package response_jenis_seller

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct Ajukan Ubah Jenis Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ResponseMasukanDataDistributor struct {
	Message string `json:"pesan_masukan_data_distributor"`
}

type ResponseEditDataDistributor struct {
	Message string `json:"pesan_edit_data_distributor"`
}

type ResponseHapusDataDistributor struct {
	Message string `json:"pesan_hapus_data_distributor"`
}

type ResponseMasukanDataBrand struct {
	Message string `json:"pesan_masukan_data_brand"`
}

type ResponseEditDataBrand struct {
	Message string `json:"pesan_edit_data_brand"`
}

type ResponseHapusDataBrand struct {
	Message string `json:"pesan_hapus_data_brand"`
}
