package seller_alamat_services

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_alamat_services_seller "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services/response_alamat_service_seller"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Alamat Gudang Seller
// Berfungsi Untuk Menulis Ke table alamat_gudang tentang alamat gudang seller tersebut, tidak ada batasan maksimal
// gudang yang boleh dilampirkan alamat nya
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TambahAlamatGudang(data PayloadTambahAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "TambahAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

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

	data.Data.ID = 0
	data.Data.IDSeller = data.IdentitasSeller.IdSeller

	if err_tambah_alamat := db.Create(&data.Data).Error; err_tambah_alamat != nil {
		log.Printf("[ERROR] Gagal menambah alamat gudang untuk seller ID %d: %v", data.IdentitasSeller.IdSeller, err_tambah_alamat)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseTambahAlamatGudang{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil ditambahkan untuk seller ID %d", data.IdentitasSeller.IdSeller)
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

func EditAlamatGudang(data PayloadEditAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "EditAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

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

	if err_edit_alamat := db.Model(&models.AlamatGudang{}).Where(models.AlamatGudang{
		ID: data.Data.ID,
	}).Updates(&data.Data).Error; err_edit_alamat != nil {
		log.Printf("[ERROR] Gagal mengedit alamat gudang ID %d untuk seller ID %d: %v", data.Data.ID, data.IdentitasSeller.IdSeller, err_edit_alamat)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseEditAlamatGudang{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
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

func HapusAlamatGudang(data PayloadHapusAlamatGudang, db *gorm.DB) *response.ResponseForm {
	services := "HapusAlamatGudang"

	_, status := data.IdentitasSeller.Validating(db)

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

	var count int64 = 0

	if err_check := db.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
		IDAlamat: data.IdGudang,
	}).Count(&count).Error; err_check != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Gagal Server Sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if count != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: fmt.Sprintf("Gagal Ganti dahulu alamat kategori mu sejumlah %v", count),
			},
		}
	}

	if err_hapus := db.Model(&models.AlamatGudang{}).Where(models.AlamatGudang{
		ID:       data.IdGudang,
		IDSeller: data.IdentitasSeller.IdSeller,
	}).Delete(&models.AlamatGudang{}).Error; err_hapus != nil {
		log.Printf("[ERROR] Gagal menghapus alamat gudang ID %d untuk seller ID %d: %v", data.IdGudang, data.IdentitasSeller.IdSeller, err_hapus)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	log.Printf("[INFO] Alamat gudang berhasil dihapus untuk seller ID %d", data.IdentitasSeller.IdSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_alamat_services_seller.ResponseHapusAlamatGudang{
			Message: "Alamat gudang berhasil dihapus.",
		},
	}
}
