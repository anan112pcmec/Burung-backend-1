package seller_transaksi_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"
func ApproveOrderTransaksi(ctx context.Context, data PayloadApproveOrder, db *gorm.DB

	entity_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/entity"
	pengiriman_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/pengiriman"
	transaksi_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/transaksi"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"

func ApproveOrderTransaksi(ctx context.Context, data PayloadApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "ApproveOrderTransaksi"

	if _, status := data.IdentitasSeller.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data seller tidak ditemukan",
		}
	}

	var dataTransaksi models.Transaksi
	if err := db.WithContext(ctx).Model(&models.Transaksi{}).Where(&models.Transaksi{
		ID:       data.IdTransaksi,
		IdSeller: data.IdentitasSeller.IdSeller,
		Status:   transaksi_enums.Dibayar,
	}).Limit(1).Take(&dataTransaksi).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if dataTransaksi.ID != data.IdTransaksi {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Message:  "Gagal data transaksi tidak valid",
		}
	}
	if dataTransaksi.BeratTotalKg == 0 {
		dataTransaksi.BeratTotalKg = 1
	}
	if dataTransaksi.IsEkspedisi {
		var id_data_pengiriman_eks int64 = 0
		if err := db.WithContext(ctx).Model(&models.PengirimanEkspedisi{}).Select("id").Where(&models.PengirimanEkspedisi{
			IdTransaksi: dataTransaksi.ID,
		}).Limit(1).Scan(&id_data_pengiriman_eks).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Message:  "Gagal server sedang sibuk coba lagi lain waktu",
			}
		}

		if id_data_pengiriman_eks != 0 {
			return &response.ResponseForm{
				Status:   http.StatusUnauthorized,
				Services: services,
				Message:  "Gagal transaksi tersebut sudah dikirim",
			}
		}

		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&models.PengirimanEkspedisi{
				IdTransaksi:       dataTransaksi.ID,
				IdAlamatGudang:    dataTransaksi.IdAlamatGudang,
				IdAlamatEkspedisi: dataTransaksi.IdAlamatEkspedisi,
				IdKurir:           nil,
				BeratBarang:       dataTransaksi.BeratTotalKg,
				KendaraanRequired: dataTransaksi.KendaraanPengiriman,
				JenisPengiriman:   dataTransaksi.JenisPengiriman,
				JarakTempuh:       dataTransaksi.JarakTempuh,
				KurirPaid:         dataTransaksi.KurirPaid,
				Status:            pengiriman_enums.PickedUp,
			}).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.Transaksi{}).Where(&models.Transaksi{
				ID: dataTransaksi.ID,
			}).Updates(&models.Transaksi{
				Status:  transaksi_enums.Diproses,
				Catatan: data.Catatan,
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
	} else {
		var id_data_pengiriman int64 = 0
		if err := db.WithContext(ctx).Model(&models.Pengiriman{}).Select("id").Where(&models.Pengiriman{
			IdTransaksi: dataTransaksi.ID,
		}).Limit(1).Scan(&id_data_pengiriman).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Message:  "Gagal server sedang sibuk coba lagi lain waktu",
			}
		}

		if id_data_pengiriman != 0 {
			return &response.ResponseForm{
				Status:   http.StatusUnauthorized,
				Services: services,
				Message:  "Gagal transaksi tersebut sudah dikirim",
			}
		}
		if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&models.Pengiriman{
				IdTransaksi:       dataTransaksi.ID,
				IdAlamatGudang:    dataTransaksi.IdAlamatGudang,
				IdAlamatPengguna:  dataTransaksi.IdAlamatPengguna,
				IdKurir:           nil,
				BeratBarang:       dataTransaksi.BeratTotalKg,
				KendaraanRequired: dataTransaksi.KendaraanPengiriman,
				JenisPengiriman:   dataTransaksi.JenisPengiriman,
				JarakTempuh:       dataTransaksi.JarakTempuh,
				KurirPaid:         dataTransaksi.KurirPaid,
				Status:            pengiriman_enums.PickedUp,
			}).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.Transaksi{}).Where(&models.Transaksi{
				ID: dataTransaksi.ID,
			}).Updates(&models.Transaksi{
				Status:  transaksi_enums.Diproses,
				Catatan: data.Catatan,
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
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Message:  "Berhasil approve transaksi",
	}
}

func UnApproveOrderTransaksi(ctx context.Context, data PayloadUnApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "UnApproveOrderTransaksi"

	if _, status := data.IdentitasSeller.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data seller tidak ditemukan",
		}
	}

	var id_data_transaksi int64 = 0
	if err := db.WithContext(ctx).Model(&models.Transaksi{}).Select("id").Where(&models.Transaksi{
		ID:       data.IdTransaksi,
		IdSeller: data.IdentitasSeller.IdSeller,
		Status:   transaksi_enums.Dibayar,
	}).Limit(1).Scan(&id_data_transaksi).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if id_data_transaksi == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data transaksi tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Model(&models.Transaksi{}).Where(&models.Transaksi{
		ID: data.IdTransaksi,
	}).Updates(&models.Transaksi{
		Status:         transaksi_enums.Dibatalkan,
		DibatalkanOleh: &entity_enums.Seller,
		Catatan:        data.Catatan,
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
