package kurir_pengiriman_services

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	kurir_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/entity/kurir"
	pengiriman_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/pengiriman"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func AktifkanBidKurir(ctx context.Context, data PayloadAktifkanBidKurir, db *gorm.DB) *response.ResponseForm {
	services := "AktifkanBidKurir"

	kurir, status := data.IdentitasKurir.Validating(ctx, db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	if kurir.StatusBid != kurir_enums.Off {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Kamu sudah melakukan bid",
		}
	}

	if !kurir.VerifiedKurir {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal harap lengkapi data minimum kurir terlebih dahulu",
		}
	}

	var id_data_kurir_bid int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		IdKurir: data.IdentitasKurir.IdKurir,
		Selesai: nil,
	}).Limit(1).Scan(&id_data_kurir_bid).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_kurir_bid != 0 {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Message:  "Gagal bid sebelumnya belum selesai",
		}
	}

	var BookPengiriman int8 = 0

	switch data.JenisPengiriman {
	case pengiriman_enums.Instant:
		BookPengiriman = 1
	case pengiriman_enums.Fast:
		BookPengiriman = 5
	case pengiriman_enums.Reguler:
		BookPengiriman = 8
	default:
		BookPengiriman = 0
	}

	if data.Mode == "manual" && data.JenisPengiriman != pengiriman_enums.Reguler {
		data.Mode = "auto"
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.BidKurirData{
			IdKurir:          data.IdentitasKurir.IdKurir,
			JenisPengiriman:  data.JenisPengiriman,
			Mode:             data.Mode,
			Provinsi:         data.Provinsi,
			Kota:             data.Kota,
			Alamat:           data.Alamat,
			Longitude:        data.Longitude,
			Latitude:         data.Latitude,
			MaxKg:            int16(data.MaxKg),
			JenisKendaraan:   kurir.TipeKendaraan,
			BookedPengiriman: int32(BookPengiriman),
			Status:           kurir_enums.Mengumpulkan,
			Dimulai:          time.Now(),
			Selesai:          nil,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Kurir{}).Where(&models.Kurir{
			ID: data.IdentitasKurir.IdKurir,
		}).Update("status_bid", kurir_enums.Idle).Error; err != nil {
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

func UpdatePosisiBidKurir(ctx context.Context, data PayloadUpdatePosisiBid, db *gorm.DB) *response.ResponseForm {
	services := "UpdatePosisiBidKurir"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var id_bid_kurir_data int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:      data.IdBidKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_bid_kurir_data).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_bid_kurir_data == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidka ditemukan",
		}
	}

	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Where(&models.BidKurirData{
		ID: data.IdBidKurir,
	}).Updates(&models.BidKurirData{
		Longitude: data.Longitude,
		Latitude:  data.Latitude,
	}).Error; err != nil {
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

func AmbilPengirimanManual(ctx context.Context, data PayloadAmbilPengirimanManual, db *gorm.DB) *response.ResponseForm {
	services := "AmbilPengirimanManual"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var data_pengiriman models.Pengiriman = models.Pengiriman{ID: 0}
	if err := db.WithContext(ctx).Model(&models.Pengiriman{}).Where(&models.Pengiriman{
		ID: data.IdPengiriman,
	}).Limit(1).Scan(&data_pengiriman).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if data_pengiriman.ID == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data pengiriman tidak ditemukan",
		}
	}

	var id_data_bid int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:      data.IdBid,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_data_bid).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_bid == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidak ditemukan",
		}
	}

	var Limit int8 = 0

	switch data_pengiriman.JenisPengiriman {
	case pengiriman_enums.Reguler:
		Limit = 10
	case pengiriman_enums.Fast:
		Limit = 5
	case pengiriman_enums.Instant:
		Limit = 1
	}

	var ids_data_bid_kurir_scheduler []int64
	if err := db.WithContext(ctx).Model(&models.BidKurirNonEksScheduler{}).Select("id").Where(&models.BidKurirNonEksScheduler{
		IdBid: data.IdBid,
	}).Limit(int(Limit)).Scan(&ids_data_bid_kurir_scheduler).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if len(ids_data_bid_kurir_scheduler) >= int(Limit) {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal kamu mencapai batas mengambil pengiriman",
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Pengiriman{}).Where(&models.Pengiriman{
			ID: data.IdPengiriman,
		}).Updates(&models.Pengiriman{
			IdKurir: &data.IdentitasKurir.IdKurir,
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.BidKurirNonEksScheduler{
			IdBid:        data.IdBid,
			IdKurir:      data.IdentitasKurir.IdKurir,
			Urutan:       int8(len(ids_data_bid_kurir_scheduler) + 1),
			IdPengiriman: data.IdPengiriman,
			Status:       "Wait",
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBid,
		}).Updates(map[string]interface{}{
			"booked_pengiriman": gorm.Expr("booked_pengiriman + ?", 1),
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
		Message:  "Berhasil",
	}
}

func LockSiapAntarBidKurir(ctx context.Context, data PayloadLockSiapAntar, db *gorm.DB) *response.ResponseForm {
	services := "LockSiapAntarBidKurir"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var id_data_bid int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:              data.IdBidKurir,
		IdKurir:         data.IdentitasKurir.IdKurir,
		JenisPengiriman: data.JenisPengiriman,
		Status:          kurir_enums.Mengumpulkan,
	}).Limit(1).Scan(&id_data_bid).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_bid == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidak ditemukan",
		}
	}

	// Oke data bid sudah ditemukan

	var LimitTake int64 = 0
	switch data.JenisPengiriman {
	case pengiriman_enums.Reguler:
		LimitTake = 10
	case pengiriman_enums.Fast:
		LimitTake = 5
	case pengiriman_enums.Instant:
		LimitTake = 1
	}

	var ids_data_bid_kurir_scheduler []int64

	if err := db.WithContext(ctx).Model(&models.BidKurirNonEksScheduler{}).Select("id").Where(&models.BidKurirNonEksScheduler{
		IdBid:   data.IdBidKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
		Status:  "Wait",
	}).Limit(int(LimitTake)).Scan(&ids_data_bid_kurir_scheduler).Error; err != nil {
		return &response.ResponseForm{
			Message: "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if len(ids_data_bid_kurir_scheduler) < 1 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal minimal 1 barang untuk di lock",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BidKurirNonEksScheduler{}).Where("id IN ?", ids_data_bid_kurir_scheduler).Update("status", "Ambil").Error; err != nil {
			return err
		}

		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBidKurir,
		}).Update("status", "Siap Antar").Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Kurir{}).Where(&models.Kurir{
			ID: data.IdentitasKurir.IdKurir,
		}).Update("status_bid", "OnDelivery").Error; err != nil {
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
		Message:  "Selamat mengantar Paket",
	}
}

func NonaktifkanBidKurir(ctx context.Context, data PayloadNonaktifkanBidKurir, db *gorm.DB) *response.ResponseForm {
	services := "NonaktifkanBidKurir"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	var id_exist_bid int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id").Where(&models.BidKurirData{
		ID:      data.IdBidKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&id_exist_bid).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_exist_bid == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal menemukan data bid",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBidKurir,
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
