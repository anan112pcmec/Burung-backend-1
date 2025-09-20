package pengguna_transaction_services

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

func CheckoutBarangUser(data PayloadCheckoutBarangCentang, db *gorm.DB) *response.ResponseForm {
	services := "CheckoutBarangUser"

	if data.IDPengguna == 0 && data.Username == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var responseData []response_transaction_pengguna.CheckoutData
	var processCheckout []models.Keranjang

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, keranjang := range data.DataCheckout {
			var jumlahStok int64
			if err := tx.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{
					IdBarangInduk: keranjang.IdBarangInduk,
					IdKategori:    keranjang.IdKategori,
					Status:        "Ready",
				}).
				Count(&jumlahStok).Error; err != nil {
				return err
			}

			var barang models.BarangInduk
			if err := tx.Select("nama_barang").
				Where("id = ?", keranjang.IdBarangInduk).
				First(&barang).Error; err != nil {
				return err
			}

			var kategori models.KategoriBarang
			if err := tx.Select("nama").
				Where("id = ?", keranjang.IdKategori).
				First(&kategori).Error; err != nil {
				return err
			}

			resp := response_transaction_pengguna.CheckoutData{
				IdBarangInduk:    keranjang.IdBarangInduk,
				IdKategoriBarang: keranjang.IdKategori,
				NamaBarang:       barang.NamaBarang,
				NamaKategori:     kategori.Nama,
				Dipesan:          int32(keranjang.Count),
			}

			if jumlahStok >= int64(keranjang.Count) {
				var varianIDs []int64
				if err := tx.Model(&models.VarianBarang{}).
					Where("id_barang_induk = ? AND id_kategori = ? AND status = ?", keranjang.IdBarangInduk, keranjang.IdKategori, "Ready").
					Limit(int(keranjang.Count)).
					Pluck("id", &varianIDs).Error; err != nil {
					resp.IdBarangInduk = keranjang.IdBarangInduk
					resp.IdKategoriBarang = keranjang.IdKategori
					resp.Message = "Coba Lagi Nanti"
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				}

				if len(varianIDs) < int(keranjang.Count) {
					shortfall := int64(keranjang.Count) - int64(len(varianIDs))
					resp.Message = fmt.Sprintf("Gagal, stok kurang %v barang", shortfall)
					resp.Status = false
					responseData = append(responseData, resp)
					continue
				}

				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Update("status", "Dipesan").Error; err != nil {
					resp.Message = "Coba Lagi Nanti"
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				}

				resp.Message = "Berhasil Siap Transaksi"
				resp.Status = true
			} else {
				shortfall := int64(keranjang.Count) - jumlahStok
				resp.Message = fmt.Sprintf("Gagal, stok kurang %v barang", shortfall)
				resp.Status = false
			}

			processCheckout = append(processCheckout, keranjang)
			responseData = append(responseData, resp)
		}
		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseDataCheckout{
				Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseDataCheckout{
			Message:      "Berhasil",
			DataResponse: responseData,
		},
	}
}

func BatalCheckoutUser(data response_transaction_pengguna.ResponseDataCheckout, db *gorm.DB) *response.ResponseForm {
	services := "BatalCheckoutKeranjang"
	var responseData []response_transaction_pengguna.CheckoutData

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, keranjang := range data.DataResponse {

			resp := response_transaction_pengguna.CheckoutData{
				IdBarangInduk:    keranjang.IdBarangInduk,
				IdKategoriBarang: keranjang.IdKategoriBarang,
				NamaBarang:       keranjang.NamaBarang,
				NamaKategori:     keranjang.NamaKategori,
				Dipesan:          keranjang.Dipesan,
			}

			var varianIDs []int64

			if err := tx.Model(&models.VarianBarang{}).
				Where(models.VarianBarang{IdBarangInduk: keranjang.IdBarangInduk, IdKategori: keranjang.IdKategoriBarang, Status: "Dipesan"}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varianIDs).Error; err != nil {
				resp.Message = "Gagal Membatalkan, Coba Lagi Nanti"
				resp.Status = false
				responseData = append(responseData, resp)
				return err
			}

			// Kalau ada ID, update balik jadi Ready
			if len(varianIDs) > 0 {
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Update("status", "Ready").Error; err != nil {
					resp.Message = "Gagal Membatalkan, Coba Lagi Nanti"
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				}

				resp.Message = "Berhasil Dibatalkan"
				resp.Status = true
			} else {
				resp.Message = "Tidak Ada Barang Dipesan Untuk Dibatalkan"
				resp.Status = false
			}

			responseData = append(responseData, resp)
		}
		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseDataCheckout{
				Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseDataCheckout{
			Message:      "Berhasil Membatalkan Checkout",
			DataResponse: responseData,
		},
	}
}
