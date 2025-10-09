package kurir_pengiriman_services

import (
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_pengiriman_services_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services/response_pengiriman_services"
)

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
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.Kredensial.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	if len(data.DataPengiriman) > MAX_PENGIRIMAN {
		log.Printf("[WARN] Kurir ID %d mencoba mengambil lebih dari %d pengiriman", data.Kredensial.IdKurir, MAX_PENGIRIMAN)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Maksimal hanya boleh mengambil 5 pengiriman di waktu bersamaan",
			},
		}
	}

	var jumlahPengirimanAktif int64
	if err := db.Model(&models.Pengiriman{}).Where(models.Pengiriman{
		IdKurir: kurir.ID,
	}).Count(&jumlahPengirimanAktif).Error; err != nil {
		log.Printf("[ERROR] Gagal mengambil jumlah pengiriman aktif untuk kurir ID %d: %v", kurir.ID, err)
		jumlahPengirimanAktif = 0
	}

	if jumlahPengirimanAktif >= MAX_PENGIRIMAN {
		log.Printf("[WARN] Kurir ID %d sudah memiliki %d pengiriman tertunda", kurir.ID, MAX_PENGIRIMAN)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Kamu sudah memiliki 5 pengiriman tertunda",
			},
		}
	}
	sisaSlot := int(MAX_PENGIRIMAN - jumlahPengirimanAktif)
	if sisaSlot <= 0 {
		log.Printf("[WARN] Tidak ada slot pengiriman tersedia untuk kurir ID %d", kurir.ID)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal, Tidak ada slot pengiriman yang tersedia",
			},
		}
	}

	if len(data.DataPengiriman) > sisaSlot {
		log.Printf("[INFO] Pengiriman yang diambil melebihi slot, hanya %d yang diproses untuk kurir ID %d", sisaSlot, kurir.ID)
		data.DataPengiriman = data.DataPengiriman[:sisaSlot]
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, pengiriman := range data.DataPengiriman {
			var status response_pengiriman_services_kurir.AmbilPengirimanDetail
			status.DataPengiriman = pengiriman

			var existingPengiriman models.Pengiriman
			_ = tx.Model(&models.Pengiriman{}).
				Where(&models.Pengiriman{
					ID:                 pengiriman.ID,
					IdTransaksi:        pengiriman.IdTransaksi,
					IdKurir:            0,
					IdAlamatPengiriman: pengiriman.IdAlamatPengiriman,
				}).Take(&existingPengiriman)

			if existingPengiriman.ID == 0 {
				log.Printf("[WARN] Pengiriman ID %d tidak ditemukan atau sudah diambil", pengiriman.ID)
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
					log.Printf("[WARN] Tipe kendaraan tidak sesuai untuk pengiriman ID %d oleh kurir ID %d", pengiriman.ID, kurir.ID)
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
					log.Printf("[ERROR] Gagal update pengiriman ID %d: %v", existingPengiriman.ID, result.Error)
					status.Status = false
					hasilPengirimanStatus = append(hasilPengirimanStatus, status)
					continue
				}
				if result.RowsAffected == 0 {
					log.Printf("[WARN] Tidak ada perubahan pada pengiriman ID %d", existingPengiriman.ID)
					status.Status = false
				} else {
					log.Printf("[INFO] Pengiriman ID %d berhasil diambil oleh kurir ID %d", existingPengiriman.ID, kurir.ID)
					status.Status = true
				}
			}
			hasilPengirimanStatus = append(hasilPengirimanStatus, status)
		}
		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal mengambil pengiriman untuk kurir ID %d: %v", kurir.ID, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
				Message: "Gagal mengambil pengiriman",
			},
		}
	}

	log.Printf("[INFO] Proses pengambilan pengiriman selesai untuk kurir ID %d", kurir.ID)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_pengiriman_services_kurir.ResponseAmbilPengiriman{
			Message:       "Berhasil",
			HasilResponse: hasilPengirimanStatus,
		},
	}
}

func UpdatePengirimanKurir(data PayloadUpdatePengiriman, db *gorm.DB) *response.ResponseForm {
	services := "UpdatePengirimanKurir"

	_, status := data.Kredensial.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.Kredensial.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseUpdatePengiriman{
				Message: "Gagal, Kredensial Kurir Tidak Valid",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		var statusP string
		_ = tx.Model(models.Pengiriman{}).Select("status").Where(models.Pengiriman{
			ID: data.DataJejakPengiriman.IdPengiriman,
		}).Take(&statusP)

		if statusP != data.StatusPengiriman {
			if err_update := tx.Model(models.Pengiriman{}).Where(models.Pengiriman{
				ID: data.DataJejakPengiriman.IdPengiriman,
			}).Limit(1).Update("status", data.StatusPengiriman).Error; err_update != nil {
				log.Printf("[ERROR] Gagal update status pengiriman ID %d: %v", data.DataJejakPengiriman.IdPengiriman, err_update)
				return err_update
			}
		}

		if err_buatjejakpengiriman := tx.Create(&data.DataJejakPengiriman).Error; err_buatjejakpengiriman != nil {
			log.Printf("[ERROR] Gagal membuat jejak pengiriman untuk pengiriman ID %d: %v", data.DataJejakPengiriman.IdPengiriman, err_buatjejakpengiriman)
			return err_buatjejakpengiriman
		}

		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal update pengiriman untuk kurir ID %d: %v", data.Kredensial.IdKurir, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_pengiriman_services_kurir.ResponseUpdatePengiriman{
				Message: "Gagal, Server Sedang Sibuk coba ulang lagi",
			},
		}
	}

	log.Printf("[INFO] Pengiriman ID %d berhasil diupdate oleh kurir ID %d", data.DataJejakPengiriman.IdPengiriman, data.Kredensial.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_pengiriman_services_kurir.ResponseUpdatePengiriman{
			Message: "Berhasil",
		},
	}
}
