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

	var SlotTersisa int8 = 0

	switch data.JenisPengiriman {
	case pengiriman_enums.Instant:
		SlotTersisa = 1
	case pengiriman_enums.Fast:
		SlotTersisa = 5
	case pengiriman_enums.Reguler:
		SlotTersisa = 8
	default:
		SlotTersisa = 0
	}

	if data.Mode == "manual" && data.JenisPengiriman != pengiriman_enums.Reguler {
		data.Mode = "auto"
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.BidKurirData{
			IdKurir:         data.IdentitasKurir.IdKurir,
			JenisPengiriman: data.JenisPengiriman,
			Mode:            data.Mode,
			Provinsi:        data.Provinsi,
			Kota:            data.Kota,
			Alamat:          data.Alamat,
			Longitude:       data.Longitude,
			Latitude:        data.Latitude,
			MaxKg:           int16(data.MaxKg),
			JenisKendaraan:  kurir.TipeKendaraan,
			IsEkspedisi:     data.IsEkspedisi,
			SlotTersisa:     int32(SlotTersisa),
			Status:          kurir_enums.Mengumpulkan,
			Dimulai:         time.Now(),
			Selesai:         nil,
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

func AmbilPengirimanNonEksManualReguler(ctx context.Context, data PayloadAmbilPengirimanNonEksManualReguler, db *gorm.DB) *response.ResponseForm {
	services := "AmbilPengirimanManualReguler"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	// Memastikan data bid ada
	var bid_data models.BidKurirData = models.BidKurirData{ID: 0}
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id", "jenis_kendaraan", "slot_tersisa", "jenis_pengiriman").Where(&models.BidKurirData{
		ID:          data.IdBid,
		IdKurir:     data.IdentitasKurir.IdKurir,
		Mode:        kurir_enums.Manual,
		IsEkspedisi: false,
	}).Limit(1).Scan(&bid_data).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if bid_data.ID == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidak ditemukan",
		}
	}

	if bid_data.SlotTersisa == 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal Slot mencapai batas",
		}
	}

	var id_same_bid_scheduler int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirNonEksScheduler{}).Select("id").Where(&models.BidKurirNonEksScheduler{
		IdPengiriman: data.IdPengiriman,
	}).Limit(1).Scan(&id_same_bid_scheduler).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_same_bid_scheduler != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal kamu sudah ambil barang itu",
		}
	}

	// Memastikan data pengiriman ada
	var jenis_kendaraan string = ""
	if err := db.WithContext(ctx).Model(&models.Pengiriman{}).Select("kendaraan_required").Where(&models.Pengiriman{
		ID:      data.IdPengiriman,
		Status:  pengiriman_enums.Waiting,
		IdKurir: nil,
	}).Limit(1).Scan(&jenis_kendaraan).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if jenis_kendaraan == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data pengiriman tidak ditemukan",
		}
	}

	if jenis_kendaraan != bid_data.JenisKendaraan {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal jenis kendaraan tidak sesuai",
		}
	}

	var max_slot int8 = 8

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			Urutan:       max_slot - int8(bid_data.SlotTersisa) + 1,
			IdPengiriman: data.IdPengiriman,
			Status:       kurir_enums.Wait,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBid,
		}).Updates(map[string]interface{}{
			"slot_tersisa": gorm.Expr("slot_tersisa - 1"),
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

	if bid_data.SlotTersisa == 1 {
		if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBid,
		}).Update("status", kurir_enums.SiapAntar).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusProcessing,
				Services: services,
				Message:  "Tunggu sebentar",
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil",
	}
}

func AmbilPengirimanEksManualReguler(ctx context.Context, data PayloadAmbilPengirimanEksManualReguler, db *gorm.DB) *response.ResponseForm {
	services := "AmbilPengirimanEksManualReguler"

	if _, status := data.IdentitasKurir.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data kurir tidak ditemukan",
		}
	}

	// Memastikan data bid ada
	var bid_data models.BidKurirData = models.BidKurirData{ID: 0}
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Select("id", "jenis_kendaraan", "slot_tersisa", "jenis_pengiriman").Where(&models.BidKurirData{
		ID:          data.IdBid,
		IdKurir:     data.IdentitasKurir.IdKurir,
		Mode:        kurir_enums.Manual,
		IsEkspedisi: true,
	}).Limit(1).Scan(&bid_data).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if bid_data.ID == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidak ditemukan",
		}
	}

	if bid_data.SlotTersisa == 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal Slot mencapai batas",
		}
	}

	// Memastikan data pengiriman ada
	var jenis_kendaraan string = ""
	if err := db.WithContext(ctx).Model(&models.PengirimanEkspedisi{}).Select("kendaraan_required").Where(&models.PengirimanEkspedisi{
		ID:     data.IdPengiriman,
		Status: pengiriman_enums.Waiting,
	}).Limit(1).Scan(&jenis_kendaraan).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if jenis_kendaraan == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data pengiriman tidak ditemukan",
		}
	}

	if jenis_kendaraan != bid_data.JenisKendaraan {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal jenis kendaraan tidak sesuai",
		}
	}

	var id_bid_scheduler int64 = 0
	if err := db.WithContext(ctx).Model(&models.BidKurirEksScheduler{}).Select("id").Where(&models.BidKurirEksScheduler{
		IdPengirimanEks: data.IdPengiriman,
	}).Limit(1).Scan(id_bid_scheduler).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_bid_scheduler != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal kamu sudah mengambil barang itu pada bid mu",
		}
	}

	var max_slot int64 = 8

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PengirimanEkspedisi{}).Where(&models.PengirimanEkspedisi{
			ID: data.IdPengiriman,
		}).Updates(&models.PengirimanEkspedisi{
			IdKurir: &data.IdentitasKurir.IdKurir,
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.BidKurirEksScheduler{
			IdBid:           data.IdBid,
			IdKurir:         data.IdentitasKurir.IdKurir,
			Urutan:          int8(max_slot) - int8(bid_data.SlotTersisa) + 1,
			IdPengirimanEks: data.IdPengiriman,
			Status:          kurir_enums.Wait,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBid,
		}).Update("slot_tersisa", gorm.Expr("slot_tersisa - 1")).Error; err != nil {
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

	if bid_data.SlotTersisa == 1 {
		if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Where(&models.BidKurirData{
			ID: data.IdBid,
		}).Update("status", kurir_enums.SiapAntar).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusProcessing,
				Services: services,
				Message:  "Tunggu sebentar",
			}
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

	var data_bid_kurir models.BidKurirData = models.BidKurirData{ID: 0}
	if err := db.WithContext(ctx).Model(&models.BidKurirData{}).Where(&models.BidKurirData{
		ID:      data.IdBidKurir,
		IdKurir: data.IdentitasKurir.IdKurir,
	}).Limit(1).Scan(&data_bid_kurir).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if data_bid_kurir.ID == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data bid tidak ditemukan",
		}
	}

	// Oke data bid sudah ditemukan

	var ids_data_bid_kurir_scheduler []int64

	if data_bid_kurir.IsEkspedisi {
		if err := db.WithContext(ctx).Model(&models.BidKurirEksScheduler{}).Select("id").Where(&models.BidKurirEksScheduler{
			IdBid:   data.IdBidKurir,
			IdKurir: data.IdentitasKurir.IdKurir,
			Status:  "Wait",
		}).Limit(8 - int(data_bid_kurir.SlotTersisa)).Scan(&ids_data_bid_kurir_scheduler).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Message:  "Gagal server sedang sibuk coba lagi lain waktu",
			}
		}
	} else {
		if err := db.WithContext(ctx).Model(&models.BidKurirNonEksScheduler{}).Select("id").Where(&models.BidKurirNonEksScheduler{
			IdBid:   data.IdBidKurir,
			IdKurir: data.IdentitasKurir.IdKurir,
			Status:  "Wait",
		}).Limit(8 - int(data_bid_kurir.SlotTersisa)).Scan(&ids_data_bid_kurir_scheduler).Error; err != nil {
			return &response.ResponseForm{
				Message: "Gagal server sedang sibuk coba lagi lain waktu",
			}
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

		if data_bid_kurir.IsEkspedisi {
			if err := tx.Model(&models.BidKurirEksScheduler{}).Where("id IN ?", ids_data_bid_kurir_scheduler).Update("status", "Ambil").Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.BidKurirNonEksScheduler{}).Where("id IN ?", ids_data_bid_kurir_scheduler).Update("status", "Ambil").Error; err != nil {
				return err
			}
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
