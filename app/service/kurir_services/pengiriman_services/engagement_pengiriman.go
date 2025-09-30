package kurir_pengiriman_services

import (
	"net/http"

	"gorm.io/gorm"
func AmbilPengirimanKurir(data PayloadAmbilPengiriman, db *gorm.DB
func AmbilPengirimanKurir(data PayloadAmbilPengiriman, db *gorm.DB

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_pengiriman_services_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services/response_pengiriman_services"

func AmbilPengirimanKurir(data PayloadAmbilPengiriman, db *gorm.DB) *response.ResponseForm {
	const MAX_PENGIRIMAN = 5
	services := "AmbilPengirimanKurir"
	var hasilPengirimanStatus []response_pengiriman_services_kurir.AmbilPengirimanDetail
	var kurir models.Kurir

	if err := db.Model(&models.Kurir{}).Where(models.Kurir{
		ID:            data.Kredensial.IdKurir,
		Username:      data.Kredensial.UsernameKurir,
		Email:         data.Kredensial.EmailKurir,
		VerifiedKurir: true,
	}).Limit(1).Take(&kurir).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	if len(data.DataPengiriman) > MAX_PENGIRIMAN {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Maksimal hanya boleh mengambil 5 pengiriman di waktu bersamaan",
			},
		}
	}

	var jumlahPengirimanAktif int64
	if err := db.Model(&models.Pengiriman{}).Where("id_kurir = ?", kurir.ID).Count(&jumlahPengirimanAktif).Error; err != nil {
		jumlahPengirimanAktif = 0
	}

	if jumlahPengirimanAktif >= MAX_PENGIRIMAN {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Kamu sudah memiliki 5 pengiriman tertunda",
			},
		}
	}
	sisaSlot := int(MAX_PENGIRIMAN - jumlahPengirimanAktif)
	if sisaSlot <= 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Tidak ada slot pengiriman yang tersedia",
			},
		}
	}

	if len(data.DataPengiriman) > sisaSlot {
		data.DataPengiriman = data.DataPengiriman[:sisaSlot]
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, pengiriman := range data.DataPengiriman {
			var status response_pengiriman_services_kurir.AmbilPengirimanDetail
			status.DataPengiriman = pengiriman

			var existingPengiriman models.Pengiriman
			_ = tx.Model(&models.Pengiriman{}).
				Where(&models.Pengiriman{
					ID:          pengiriman.ID,
					IdTransaksi: pengiriman.IdTransaksi,
					IdKurir:     0,
					IdAlamat:    pengiriman.IdAlamat,
				}).Take(&existingPengiriman)

			if existingPengiriman.ID == 0 {
				status.Status = false
				hasilPengirimanStatus = append(hasilPengirimanStatus, status)
				continue
			} else {
				var kendaraanKurir string

				_ = tx.Model(&models.Kurir{}).
					Select("tipe_kendaraan").
					Where(models.Kurir{
						ID: existingPengiriman.IdKurir,
					}).Take(&kendaraanKurir)

				if kendaraanKurir != pengiriman.Layanan {
					status.Status = false
					hasilPengirimanStatus = append(hasilPengirimanStatus, status)
					continue
				}

				result := tx.Model(&models.Pengiriman{}).
					Where(models.Pengiriman{
						ID: existingPengiriman.ID,
					}).
					Update("id_kurir", data.Kredensial.IdKurir)

				if result.Error != nil {
					status.Status = false
					hasilPengirimanStatus = append(hasilPengirimanStatus, status)
					continue
				}
				if result.RowsAffected == 0 {
					status.Status = false
				} else {
					status.Status = true
				}
			}

			hasilPengirimanStatus = append(hasilPengirimanStatus, status)
		}
		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal mengambil pengiriman",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
			Message:       "Berhasil",
			HasilResponse: hasilPengirimanStatus,
		},
	}
}
