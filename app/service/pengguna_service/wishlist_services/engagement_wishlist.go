package pengguna_wishlist_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/wishlist_services/response_wishlist_services_pengguna"
)

func TambahBarangKeWishlist(ctx context.Context, data PayloadTambahBarangKeWishlist, db *gorm.DB) *response.ResponseForm {
	services := "TambahBarangKeWishlist"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseTambahBarangKeWishlist{
				Message: "Gagal data seller tidak valid",
			},
		}
	}

	var id_data_wishlist = 0
	if err := db.WithContext(ctx).Model(&models.Wishlist{}).Select("id").Where(&models.Wishlist{
		IdPengguna:    data.IdentitasPengguna.ID,
		IdBarangInduk: data.IdBarangInduk,
	}).Limit(1).Scan(&id_data_wishlist).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseTambahBarangKeWishlist{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_wishlist != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseTambahBarangKeWishlist{
				Message: "Gagal kamu sudah memasukan barang itu ke dalam wishlist",
			},
		}
	}

	if err := db.WithContext(ctx).Create(&models.Wishlist{
		IdPengguna:    data.IdentitasPengguna.ID,
		IdBarangInduk: data.IdBarangInduk,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseTambahBarangKeWishlist{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_wishlist_services_pengguna.ResponseTambahBarangKeWishlist{
			Message: "Berhasil",
		},
	}
}

func HapusBarangDariWishlist(ctx context.Context, data PayloadHapusBarangDariWishlist, db *gorm.DB) *response.ResponseForm {
	services := "HapusBarangDariWishlist"

	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseHapusBarangDariWishlist{
				Message: "Gagal data seller tidak valid",
			},
		}
	}

	var id_data_wishlist int64 = 0
	if err := db.WithContext(ctx).Model(&models.Wishlist{}).Select("id").Where(&models.Wishlist{
		ID:         data.IdWishlist,
		IdPengguna: data.IdentitasPengguna.ID,
	}).Limit(1).Scan(&id_data_wishlist).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseHapusBarangDariWishlist{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	if id_data_wishlist == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseHapusBarangDariWishlist{
				Message: "Gagal data wishlist tidak ditemukan",
			},
		}
	}

	if err := db.WithContext(ctx).Model(&models.Wishlist{}).Where(&models.Wishlist{
		ID: data.IdWishlist,
	}).Delete(&models.Wishlist{}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_wishlist_services_pengguna.ResponseHapusBarangDariWishlist{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_wishlist_services_pengguna.ResponseHapusBarangDariWishlist{
			Message: "Berhasil",
		},
	}
}
