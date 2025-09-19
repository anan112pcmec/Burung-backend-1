package pengguna_transaction_services

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"func CheckoutBarangUser(data PayloadCheckoutBarangCentang, db *gorm.DB) *response.ResponseForm {
	services := "CheckoutBarangUser"

	if data.IDPengguna == 0 && data.Username == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var responseData []response_transaction_pengguna.CheckoutData
	var processCheckout []models.Keranjang

	for _, keranjang := range data.DataCheckout {
		// cek stok
		var jumlahStok int64
		if err := db.Model(&models.VarianBarang{}).
			Where(&models.VarianBarang{
				IdBarangInduk: keranjang.IdBarangInduk,
				IdKategori:    keranjang.IdKategori,
				Status:        "Ready",
			}).
			Count(&jumlahStok).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_transaction_pengguna.ResponseDataCheckout{
					Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
				},
			}
		}

		// ambil nama barang & kategori
		var barang models.BarangInduk
		if err := db.Select("nama_barang").
			Where("id = ?", keranjang.IdBarangInduk).
			First(&barang).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_transaction_pengguna.ResponseDataCheckout{
					Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
				},
			}
		}

		var kategori models.KategoriBarang
		if err := db.Select("nama").
			Where("id = ?", keranjang.IdKategori).
			First(&kategori).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_transaction_pengguna.ResponseDataCheckout{
					Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
				},
			}
		}

		resp := response_transaction_pengguna.CheckoutData{
			NamaBarang:   barang.NamaBarang,
			NamaKategori: kategori.Nama,
			Dipesan:      int32(keranjang.Count),
		}

		if jumlahStok >= int64(keranjang.Count) {
			resp.Message = "Berhasil Siap Transaksi"
			resp.Status = true
		} else {
			shortfall := int64(keranjang.Count) - jumlahStok
			resp.Message = fmt.Sprintf("Gagal Barang Kelebihan Dari Stok Yang ada sebanyak %v barang", shortfall)
			resp.Status = false
		}

		processCheckout = append(processCheckout, keranjang)
		responseData = append(responseData, resp)
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
