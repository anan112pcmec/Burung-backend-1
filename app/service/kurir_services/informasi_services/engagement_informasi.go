package kurir_informasi_services

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_informasi_services_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/informasi_services/response_informasi_services"
)

func AjukanInformasiKendaraan(data PayloadInformasiDataKendaraan, db *gorm.DB) *response.ResponseForm {
	services := "AjukanInformasiKendaraanKurir"

	if !data.DataIdentitasKurir.Validate() {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Kredensial Kosong",
			},
		}
	}

	var kurir int64 = 0
	if validasi_kurir := db.Model(models.Kurir{}).Select("id").Where(models.Kurir{
		ID:       data.DataIdentitasKurir.IDKurir,
		Username: data.DataIdentitasKurir.Username,
		Email:    data.DataIdentitasKurir.Email,
	}).Take(&kurir).Error; validasi_kurir != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	if kurir == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	var id_pengajuan int64 = 0
	_ = db.Model(models.InformasiKendaraanKurir{}).Select("id").Where(models.InformasiKendaraanKurir{
		IDkurir: data.DataIdentitasKurir.IDKurir,
	}).Take(&id_pengajuan)

	if id_pengajuan != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Tunggu Pengajuan Sebelum nya ditindak kami",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKendaraan.StatusPerizinan = "Pending"
		data.DataInformasiKendaraan.ID = 0
		if err_ajukan := tx.Create(&data.DataInformasiKendaraan).Error; err_ajukan != nil {
			return err_ajukan
		}

		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Server sedang sibuk coba lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
			Message: "Berhasil",
		},
	}
}

func AjukanInformasiKurir(data PayloadInformasiDataKurir, db *gorm.DB) *response.ResponseForm {
	services := "AjukanInformasiKurir"
	if !data.DataIdentitasKurir.Validate() {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, Kredensial Tidak Ditemukan",
			},
		}
	}

	var kurir int64 = 0
	if validasi_kurir := db.Model(models.Kurir{}).Select("id").Where(models.Kurir{
		ID:       data.DataIdentitasKurir.IDKurir,
		Username: data.DataIdentitasKurir.Username,
		Email:    data.DataIdentitasKurir.Email,
	}).Take(&kurir).Error; validasi_kurir != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	if kurir == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
				Message: "Gagal, Kredensial Tidak Valid",
			},
		}
	}

	var id_pengajuan int64 = 0
	_ = db.Model(models.InformasiKurir{}).Select("id").Where(models.InformasiKendaraanKurir{
		IDkurir: data.DataIdentitasKurir.IDKurir,
	}).Take(&id_pengajuan)

	if id_pengajuan != 0 {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Tunggu Pengajuan Sebelum nya ditindak kami",
			},
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		data.DataInformasiKurir.StatusPerizinan = "Pending"
		data.DataInformasiKurir.ID = 0
		if err_ajukan := tx.Create(&data.DataInformasiKurir).Error; err_ajukan != nil {
			return err_ajukan
		}
		return nil
	}); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_informasi_services_kurir.ResponseAjukanInformasiKendaraan{
				Message: "Gagal, Server sedang sibuk coba lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_informasi_services_kurir.ResponseAjukanInformasiKurir{
			Message: "Berhasil",
		},
	}
}
