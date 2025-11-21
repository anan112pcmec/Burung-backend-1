package seller_alamat_services

import (
	"context"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_alamat_services_seller "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services/response_alamat_service_seller"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Alamat Gudang Seller
// Berfungsi Untuk Menulis Ke table alamat_gudang tentang alamat gudang seller tersebut, tidak ada batasan maksimal
// gudang yang boleh dilampirkan alamat nya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahAlamatGudang(ctx context.Context, data PayloadTambahAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudang"

	_, status := data.IdentitasSeller.Validating(ctx, db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Kredensial seller tidak valid.",
			},
		}
	}

	var id_data_alamat int64 = 0
	if err := db.WithContext(ctx).Model(&models.AlamatGudang{}).Select("id").Where(&models.AlamatGudang{
		IDSeller:   data.IdentitasSeller.IdSeller,
		NamaAlamat: data.NamaAlamat,
	}).Limit(1).Scan(&id_data_alamat).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Gagal Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_alamat != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Gagal Kamu sudah memiliki alamat dengan nama yang sama",
			},
		}
	}

	helper.SanitasiKoordinat(&data.Latitude, &data.Longitude)

	if err := db.WithContext(ctx).Create(&models.AlamatGudang{
		IDSeller:        data.IdentitasSeller.IdSeller,
		PanggilanAlamat: data.PanggilanAlamat,
		NomorTelephone:  data.NomorTelefon,
		NamaAlamat:      data.NamaAlamat,
		Provinsi:        data.Provinsi,
		Kota:            data.Kota,
		KodePos:         data.KodePos,
		KodeNegara:      data.KodeNegara,
		Deskripsi:       data.Deskripsi,
		Longitude:       data.Longitude,
		Latitude:        data.Latitude,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
			Message: "Alamat gudang berhasil ditambahkan.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Edit Alamat Gudang
// Berfungsi Untuk Seller manakala mereka ingin mengedit gudang mereka entah perubahan titik, nama dll
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EditAlamatGudang(ctx context.Context, data PayloadEditAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "EditAlamatGudang"

	_, status := data.IdentitasSeller.Validating(ctx, db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Kredensial seller tidak valid.",
			},
		}
	}

	var id_data_alamat int64 = 0
	if err := db.WithContext(ctx).Model(&models.AlamatGudang{}).Select("id").Where(&models.AlamatGudang{
		ID:       data.IdAlamatGudang,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Limit(1).Scan(&id_data_alamat).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_alamat == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Gagal Data alamat tidak valid",
			},
		}
	}

	var idDataTransaksi int64 = 0

	if err := db.WithContext(ctx).
		Model(&models.Transaksi{}).
		Select("id").
		Where("id_alamat_gudang = ? AND status != ?", data.IdAlamatGudang, "Selesai").
		Limit(1).
		Scan(&idDataTransaksi).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal, server sedang sibuk. Coba lagi lain waktu",
		}
	}

	// Jika ada transaksi yang menggunakan alamat ini
	if idDataTransaksi != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal, alamat sedang digunakan sebagai acuan transaksi",
		}
	}

	helper.SanitasiKoordinat(&data.Latitude, &data.Longitude)

	if err := db.WithContext(ctx).Model(&models.AlamatGudang{}).Where(&models.AlamatGudang{
		ID: data.IdAlamatGudang,
	}).Updates(&models.AlamatGudang{
		PanggilanAlamat: data.PanggilanAlamat,
		NomorTelephone:  data.NomorTelefon,
		NamaAlamat:      data.NamaAlamat,
		Provinsi:        data.Provinsi,
		Kota:            data.Kota,
		KodePos:         data.KodePos,
		KodeNegara:      data.KodeNegara,
		Deskripsi:       data.Deskripsi,
		Longitude:       data.Longitude,
		Latitude:        data.Latitude,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil diubah untuk seller ID %d", data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
			Message: "Alamat gudang berhasil diubah.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Hapus Alamat Gudang Seller
// Berfungsi Untuk Menghapus Suatu Alamat Gudang Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func HapusAlamatGudang(ctx context.Context, data PayloadHapusAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatGudang"

	_, status := data.IdentitasSeller.Validating(ctx, db)

	if !status {
		log.Printf("[WARN] Kredensial seller tidak valid untuk ID %d", data.IdentitasSeller.IdSeller)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Kredensial seller tidak valid.",
			},
		}
	}

	var id_data_alamat int64 = 0

	if err := db.WithContext(ctx).Model(&models.AlamatGudang{}).Select("id").Where(&models.AlamatGudang{
		ID:       data.IdGudang,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Limit(1).Scan(&id_data_alamat).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_alamat == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal masukan data alamat tidak valid",
			},
		}
	}

	var total int64 = 0
	if err := db.WithContext(ctx).Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
		IDAlamat: data.IdGudang,
	}).Count(&total).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if total != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal kamu tidak bisa menghapus alamat ini karna masih digunakan oleh beberapa barangmu alihkan terlebih dahulu",
			},
		}
	}

	var idDataTransaksi int64 = 0

	if err := db.WithContext(ctx).
		Model(&models.Transaksi{}).
		Select("id").
		Where("id_alamat_gudang = ? AND status != ?", data.IdGudang, "Selesai").
		Limit(1).
		Scan(&idDataTransaksi).Error; err != nil {

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal, server sedang sibuk. Coba lagi lain waktu",
		}
	}

	// Jika ada transaksi yang menggunakan alamat ini
	if idDataTransaksi != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal, alamat sedang digunakan sebagai acuan transaksi",
		}
	}

	if err_hapus := db.WithContext(ctx).Model(&models.AlamatGudang{}).Where(models.AlamatGudang{
		ID:       data.IdGudang,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Delete(&models.AlamatGudang{}).Error; err_hapus != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
			Message: "Alamat gudang berhasil dihapus.",
		},
	}
}
