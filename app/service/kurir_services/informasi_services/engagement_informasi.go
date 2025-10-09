package kurir_informasi_services

import (
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_informasi_services_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/informasi_services/response_informasi_services"
)

func AjukanInformasiKendaraan(data PayloadInformasiDataKendaraan, db *gorm.DB) *response.ResponseForm {
	services := "AjukanInformasiKendaraanKurir"

	_, status := data.DataIdentitasKurir.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, kredensial tidak valid.",
			},
		}
	}

	var id_pengajuan int64 = 0
	_ = db.Model(models.InformasiKendaraanKurir{}).Select("id").Where(models.InformasiKendaraanKurir{
		IDkurir: data.DataIdentitasKurir.IdKurir,
	}).Take(&id_pengajuan)

	if id_pengajuan != 0 {
		log.Printf("[WARN] Sudah ada pengajuan kendaraan yang belum diproses untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, tunggu pengajuan sebelumnya ditindak kami.",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKendaraan.StatusPerizinan = "Pending"
		data.DataInformasiKendaraan.ID = 0
		if err_ajukan := tx.Create(&data.DataInformasiKendaraan).Error; err_ajukan != nil {
			log.Printf("[ERROR] Gagal mengajukan informasi kendaraan untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err_ajukan)
			return err_ajukan
		}
		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal transaksi pengajuan informasi kendaraan untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Pengajuan informasi kendaraan berhasil untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
			Message: "Berhasil.",
		},
	}
}

func EditInformasiKendaraan(data PayloadEditInformasiDataKendaraan, db *gorm.DB) *response.ResponseForm {
	services := "EditInformasiKendaraan"

	_, status := data.DataIdentitasKurir.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseEditInformasiKendaraan{
				Message: "Gagal, kredensial tidak valid.",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKendaraan.StatusPerizinan = "Pending"
		if err_updateInformasi := tx.Model(models.InformasiKendaraanKurir{}).Where(models.InformasiKendaraanKurir{
			ID:      data.DataInformasiKendaraan.ID,
			IDkurir: data.DataIdentitasKurir.IdKurir,
		}).Limit(1).Updates(&data.DataInformasiKendaraan).Error; err_updateInformasi != nil {
			log.Printf("[ERROR] Gagal mengedit informasi kendaraan ID %d untuk kurir ID %d: %v", data.DataInformasiKendaraan.ID, data.DataIdentitasKurir.IdKurir, err_updateInformasi)
			return err_updateInformasi
		}
		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal transaksi edit informasi kendaraan untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseEditInformasiKendaraan{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Edit informasi kendaraan berhasil untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseEditInformasiKendaraan{
			Message: "Berhasil.",
		},
	}

}

func AjukanInformasiKurir(data PayloadInformasiDataKurir, db *gorm.DB) *response.ResponseForm {
	services := "AjukanInformasiKurir"

	_, status := data.DataIdentitasKurir.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, kredensial tidak valid.",
			},
		}
	}

	var id_pengajuan int64 = 0
	_ = db.Model(models.InformasiKurir{}).Select("id").Where(models.InformasiKurir{
		IDkurir: data.DataIdentitasKurir.IdKurir,
	}).Take(&id_pengajuan)

	if id_pengajuan != 0 {
		log.Printf("[WARN] Sudah ada pengajuan data kurir yang belum diproses untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, tunggu pengajuan sebelumnya ditindak kami.",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKurir.StatusPerizinan = "Pending"
		data.DataInformasiKurir.ID = 0
		if err_ajukan := tx.Create(&data.DataInformasiKurir).Error; err_ajukan != nil {
			log.Printf("[ERROR] Gagal mengajukan data kurir untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err_ajukan)
			return err_ajukan
		}
		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal transaksi pengajuan data kurir untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Pengajuan data kurir berhasil untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
			Message: "Berhasil.",
		},
	}
}

func EditInformasiKurir(data PayloadEditInformasiDataKurir, db *gorm.DB) *response.ResponseForm {
	services := "EditInformasiKurir"

	_, status := data.DataIdentitasKurir.Validating(db)

	if !status {
		log.Printf("[WARN] Kredensial kurir tidak valid untuk ID %d", data.DataIdentitasKurir.IdKurir)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseEditInformasiKurir{
				Message: "Gagal, kredensial tidak valid.",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKurir.StatusPerizinan = "Pending"
		if err_edit_informasi := tx.Model(models.InformasiKurir{}).Where(models.InformasiKurir{
			ID:      data.DataInformasiKurir.ID,
			IDkurir: data.DataIdentitasKurir.IdKurir,
		}).Updates(&data.DataInformasiKurir).Error; err_edit_informasi != nil {
			log.Printf("[ERROR] Gagal mengedit data kurir ID %d untuk kurir ID %d: %v", data.DataInformasiKurir.ID, data.DataIdentitasKurir.IdKurir, err_edit_informasi)
			return err_edit_informasi
		}
		return nil
	}); err != nil {
		log.Printf("[ERROR] Gagal transaksi edit data kurir untuk kurir ID %d: %v", data.DataIdentitasKurir.IdKurir, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseEditInformasiKurir{
				Message: "Gagal, server sedang sibuk. Coba lagi lain waktu.",
			},
		}
	}

	log.Printf("[INFO] Edit data kurir berhasil untuk kurir ID %d", data.DataIdentitasKurir.IdKurir)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseEditInformasiKurir{
			Message: "Berhasil.",
		},
	}
}
