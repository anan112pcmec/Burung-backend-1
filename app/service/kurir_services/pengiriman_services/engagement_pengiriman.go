package kurir_pengiriman_services

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	kurir_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/entity/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func AktifkanBidKurir(ctx context.Context, data PayloadAktifkanBid, db *gorm.DB) *response.ResponseForm {
	services := "AktifkanBidKurir"

	if data_kurir, status := data.IdentitasKurir.Validating(ctx, db); !status {
		if !data_kurir.VerifiedKurir {
			return &response.ResponseForm{
				Status:   http.StatusUnauthorized,
				Services: services,
				Message:  "Gagal kamu belum melengkapi data kurir",
			}
		}
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var id_data_bid_kurir int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_data_bid_kurir).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_bid_kurir != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal maksimal hanya melakukan 1 bid",
		}
	}

	var jenis_kendaraan string = ""
	if err := db.WithContext(ctx).Model(&models.InformasiKendaraanKurir{}).Select("jenis_kendaraan").Where(&models.InformasiKendaraanKurir{
		IDkurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&jenis_kendaraan).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if jenis_kendaraan == "" {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal data Kendaraan mu tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&models.Kurir{}).Where(&models.Kurir{
			ID: data.IdentitasKurir.IdKurir,
		}).Update("status_bid", kurir_enums.Idle).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.BidKurirData{
			IdKurir:          data.IdentitasKurir.IdKurir,
			Mode:             data.Mode,
			Alamat:           data.Alamat,
			Longitude:        data.Longitude,
			Latitude:         data.Latitude,
			MaxJarak:         data.MaxJarak,
			MaxRadius:        data.MaxRadius,
			MaxKg:            data.MaxKg,
			BookedPengiriman: 0,
			Dimulai:          time.Now(),
			JenisKendaraan:   jenis_kendaraan,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil Menyalakan Bid",
	}
}

func UpdateLokasiBidKurir(ctx context.Context, data PayloadUpdateLokasiBidKurir, db *gorm.DB) *response.ResponseForm {
	services := "UpdateLokasiBidKurir"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var id_data_bid_kurir int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:      data.IdBidDataKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_data_bid_kurir).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_bid_kurir == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Data bid tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Where(&models.BidKurirData{
		ID: data.IdBidDataKurir,
	}).Updates(&models.BidKurirData{
		Longitude: data.Longitude,
		Latitude:  data.Latitude,
		Alamat:    data.Alamat,
	}).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}

func NonaktifBidKurir(ctx context.Context, data PayloadNonaktifkanBid, db *gorm.DB) *response.ResponseForm {
	services := "NonaktifkanBidKurir"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal server",
		}
	}

	var id_data_bid_kurir int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:      data.IdBidDataKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_data_bid_kurir).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_bid_kurir == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Data bid tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBidDataKurir,
		}).Update("berakhir", time.Now()).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBidDataKurir,
		}).Delete(&models.BidKurirData{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Kurir{}).Where(&models.Kurir{
			ID: data.IdentitasKurir.IdKurir,
		}).Update("status_bid", kurir_enums.Off).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}

func AmbilPengiriman(ctx context.Context, data PayloadAmbilPengiriman, db *gorm.DB) *response.ResponseForm {
	services := "AmbilPengiriman"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}
