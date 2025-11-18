package seller_transaksi_services

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	transaksi_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/transaksi"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"

)

func ApproveOrderTransaksi(ctx context.Context, data PayloadApproveOrder, db *gorm.DB) *response.ResponseForm {
	services := "ApproverOrderTransaksi"

	if _, status := data.IdentitasSeller.Validating(ctx, db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data seller tidak ditemukan",
		}
	}

	var data_transaksi models.Transaksi = models.Transaksi{
		ID: 0,
	}

	if err := db.WithContext(ctx).Model(&models.Transaksi{}).Where(&models.Transaksi{
		ID:       data.IdTransaksi,
		IdSeller: data.IdentitasSeller.IdSeller,
	}).Limit(1).Scan(&data_transaksi).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}

	if data.IdTransaksi == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Message:  "Gagal data transaksi tidak ditemukan",
		}
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Transaksi{}).Where(&models.Transaksi{
			ID: data.IdTransaksi,
		}).Update("status", transaksi_enums.Diproses).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.Pengiriman{
			IdTransaksi: data_transaksi.ID,
			IdAlamatPengambilan: data_transaksi.IdAlamatPengguna,
		})

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
