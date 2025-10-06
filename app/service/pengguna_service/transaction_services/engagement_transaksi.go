package pengguna_transaction_services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

// ////////////////////////////////////////////////////////////////////////////////////
// CHECKOUT
// ////////////////////////////////////////////////////////////////////////////////////

func CheckoutBarangUser(data PayloadCheckoutBarangCentang, db *gorm.DB) *response.ResponseForm {
	services := "CheckoutBarangUser"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var firstSellerID int64 = 0
	for i, keranjang := range data.DataCheckout {
		if i == 0 {
			firstSellerID = int64(keranjang.IdSeller)
		} else {
			if keranjang.IdSeller != int32(firstSellerID) {
				return &response.ResponseForm{
					Status:   http.StatusBadRequest,
					Services: services,
					Payload: response_transaction_pengguna.ResponseDataCheckout{
						Message: "Gagal, semua barang harus dari seller yang sama",
					},
				}
			}
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
			if err := tx.Select("nama_barang", "id_seller", "jenis_barang").
				Where("id = ?", keranjang.IdBarangInduk).
				First(&barang).Error; err != nil {
				return err
			}

			var kategori models.KategoriBarang
			if err := tx.Select("nama", "harga", "stok").
				Where("id = ?", keranjang.IdKategori).
				First(&kategori).Error; err != nil {
				return err
			}

			var nama_seller string
			if err_nama_seller := tx.Model(models.Seller{}).
				Select("nama").
				Where(models.Seller{ID: barang.SellerID}).
				First(&nama_seller).Error; err_nama_seller != nil {
				return err_nama_seller
			}

			resp := response_transaction_pengguna.CheckoutData{
				IDUser:           data.IdentitasPengguna.ID,
				IDSeller:         keranjang.IdSeller,
				NamaSeller:       nama_seller,
				JenisBarang:      barang.JenisBarang,
				IdBarangInduk:    keranjang.IdBarangInduk,
				IdKategoriBarang: keranjang.IdKategori,
				HargaKategori:    kategori.Harga,
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
					Updates(map[string]interface{}{
						"status":        "Dipesan",
						"hold_by":       data.IdentitasPengguna.ID,
						"holder_entity": "Pengguna",
					}).Error; err != nil {
					resp.Message = "Coba Lagi Nanti"
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				} else {

					var stok_saat_ini int64 = 0
					_ = db.Model(&models.KategoriBarang{}).Select("stok").Where(&models.KategoriBarang{
						ID: keranjang.IdKategori,
					}).Take(&stok_saat_ini)
					_ = db.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{ID: keranjang.IdKategori}).Updates(map[string]interface{}{
						"stok": int32(stok_saat_ini) - int32(keranjang.Count),
					})
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
			LayananPengiriman: response_transaction_pengguna.LayananPengiriman{
				JenisLayananKurir: data.JenisLayananKurir,
			},
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
				Where(models.VarianBarang{IdBarangInduk: keranjang.IdBarangInduk, IdKategori: keranjang.IdKategoriBarang, Status: "Dipesan", HoldBy: keranjang.IDUser}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varianIDs).Error; err != nil {
				resp.Message = "Gagal Membatalkan, Coba Lagi Nanti"
				resp.Status = false
				responseData = append(responseData, resp)
				return err
			}

			if len(varianIDs) > 0 {
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Updates(map[string]interface{}{
						"status":        "Ready",
						"hold_by":       0,
						"holder_entity": "",
					}).Error; err != nil {
					resp.Message = "Gagal Membatalkan, Coba Lagi Nanti"
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				} else {
					var stok_saat_ini int64 = 0
					_ = db.Model(&models.KategoriBarang{}).Select("stok").Where(&models.KategoriBarang{
						ID: keranjang.IdKategoriBarang,
					}).Take(&stok_saat_ini)

					_ = db.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
						ID: keranjang.IdKategoriBarang,
					}).Updates(&models.KategoriBarang{
						Stok: keranjang.Dipesan + int32(stok_saat_ini),
					})
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

// ////////////////////////////////////////////////////////////////////////////////////
// TRANSAKSI
// ////////////////////////////////////////////////////////////////////////////////////

func FormattingTransaksi(user models.Pengguna, alamat models.AlamatPengguna, data response_transaction_pengguna.ResponseDataCheckout, db *gorm.DB, PaymentMethod, jenis_layanan string) (*response.ResponseForm, *snap.Request) {
	services := "ValidateTransaksi"
	if user.ID == 0 && user.Username == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}, nil
	}

	var valid int64 = 0
	if checkuser := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: user.ID, Username: user.Username, Email: user.Email}).Count(&valid).Limit(1).Error; checkuser != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}, nil
	}

	if valid == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}, nil
	}

	if alamat.NamaAlamat == "" && alamat.PanggilanAlamat == "" && alamat.NomorTelephone == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}, nil
	}

	var PaymentCode string
	var err_payment error

	maxRetry := 10
	for i := 0; i < maxRetry; i++ {
		PaymentCode, err_payment = helper.GenerateAutoPaymentId(db)
		if err_payment == nil {
			break
		}
	}

	if err_payment != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}, nil
	}

	AlamatPengguna := midtrans.CustomerAddress{
		Address:     alamat.NamaAlamat,
		City:        alamat.Kota,
		Postcode:    alamat.KodePos,
		CountryCode: alamat.KodeNegara,
	}

	items, TotalHarga := helper.GenerateItemDetail(data, db, jenis_layanan)

	var PM []snap.SnapPaymentType

	if PaymentMethod == "VA" {
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeBCAVA,
			snap.PaymentTypeBNIVA,
			snap.PaymentTypeBRIVA,
			snap.PaymentTypePermataVA,
		}
	}

	if PaymentMethod == "E-Wallet" {
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeGopay,
			snap.PaymentTypeShopeepay,
		}
	}

	if PaymentMethod == "Gerai" {
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeIndomaret,
			snap.PaymentTypeAlfamart,
		}
	}

	if PaymentMethod == "Debit" {
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeBCAKlikpay,
			snap.PaymentTypeBRIEpay,
		}
	}

	SnapReqeust := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  PaymentCode,
			GrossAmt: TotalHarga,
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName:    "Username : " + user.Username,
			LName:    "Nama : " + user.Nama,
			Email:    user.Email,
			Phone:    alamat.NomorTelephone,
			BillAddr: &AlamatPengguna,
			ShipAddr: &AlamatPengguna,
		},
		Items:           &items,
		EnabledPayments: PM,
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}, SnapReqeust
}

func ValidateTransaksi(snapReq *snap.Request) (*snap.Response, *response.ResponseForm) {
	services := "ProsesTransaksiDenganSDK"

	var s snap.Client
	s.New("Mid-server-7wpABbBW_WURdLxcxc5bX5eb", midtrans.Sandbox)

	snapResp, err := s.CreateTransaction(snapReq)
	if err != nil {
		return nil, &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	return snapResp, &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}
}

func SnapTransaksi(data PayloadSnapTransaksiRequest, db *gorm.DB) *response.ResponseForm {
	services := "SnapTransaksiUser"

	if data.UserInformation.ID == 0 && data.UserInformation.Username == "" && data.UserInformation.Nama == "" && data.UserInformation.Email == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.AlamatInformation.NamaAlamat == "" && data.AlamatInformation.KodeNegara == "" && data.AlamatInformation.IDPengguna != data.AlamatInformation.ID && data.AlamatInformation.NomorTelephone == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	SnapErr, SnapReq := FormattingTransaksi(data.UserInformation, data.AlamatInformation, data.DataCheckout, db, data.PaymentMethod, data.DataCheckout.LayananPengiriman.JenisLayananKurir)
	if SnapErr.Status != http.StatusOK {
		return &response.ResponseForm{
			Status:   SnapErr.Status,
			Services: services,
		}
	}

	SnapResponse, SnapResponseErr := ValidateTransaksi(SnapReq)
	if SnapResponseErr.Status != http.StatusOK {
		return &response.ResponseForm{
			Status:   SnapErr.Status,
			Services: services,
		}
	}

	return &response.ResponseForm{
		Status:   SnapErr.Status,
		Services: services,
		Payload: response_transaction_pengguna.SnapTransaksi{
			SnapTransaksi: &snap.Response{
				Token:       SnapResponse.Token,
				RedirectURL: "/",
				StatusCode:  "Berhasil",
			},
			DataCheckout: data.DataCheckout.DataResponse,
			DataAlamat:   data.AlamatInformation,
		},
	}
}

func PendingTransaksi(ctx context.Context, data PayloadPendingTransaksi, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PendingTransaksi"

	if _, status := data.IdentitasPengguna.Validating(db); !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePendingTransaksi{
				Message: "Gagal Kredensial User Tidak Valid",
			},
		}
	}

	key := fmt.Sprintf("transaction_pengguna_pending_id:%v:transaction_code:%s",
		data.IdentitasPengguna.ID, data.DataPending.OrderId)

	fields := map[string]interface{}{
		"finish_redirect_url": data.DataPending.FinishRedirectUrl,
		"fraud_status":        data.DataPending.FraudStatus,
		"gross_amount":        data.DataPending.GrossAmout,
		"order_id":            data.DataPending.OrderId,
		"payment_type":        data.DataPending.PaymentType,
		"status_code":         data.DataPending.StatusCode,
		"status_message":      data.DataPending.StatusMessage,
		"transaction_id":      data.DataPending.TransactionId,
		"transaction_status":  data.DataPending.TransactionStatus,
		"transaction_time":    data.DataPending.TransactionTime,
	}

	if err := rds.HSet(ctx, key, fields).Err(); err != nil {
		fmt.Println("⚠️ Gagal menyimpan ke Redis:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePendingTransaksi{
				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponsePendingTransaksi{
			Message: "Berhasil",
		},
	}
}

func BatalTransaksi(data response_transaction_pengguna.SnapTransaksi, db *gorm.DB) *response.ResponseForm {
	services := "BatalTransaksi"

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, keranjang := range data.DataCheckout {

			var varianIDs []int64

			if err := tx.Model(&models.VarianBarang{}).
				Where(models.VarianBarang{IdBarangInduk: keranjang.IdBarangInduk, IdKategori: keranjang.IdKategoriBarang, Status: "Dipesan", HoldBy: keranjang.IDUser}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varianIDs).Error; err != nil {
				return err
			}

			if len(varianIDs) > 0 {
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Updates(map[string]interface{}{
						"status":        "Ready",
						"hold_by":       0,
						"holder_entity": "",
					}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseBatalTransaksi{
				Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseBatalTransaksi{
			Message: "Berhasil",
		},
	}
}

func LockTransaksi(data PayloadLockTransaksi, db *gorm.DB) *response.ResponseForm {
	services := "LockTransaksi"

	for _, keranjang := range data.DataHold {
		if keranjang.IDSeller == 0 && keranjang.IDUser == 0 && keranjang.IdBarangInduk == 0 {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
			}
		}
	}

	if data.PaymentResult.OrderId == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		grossFloat, err := strconv.ParseFloat(data.PaymentResult.GrossAmount, 64)
		if err != nil {
			return fmt.Errorf("invalid gross amount format: %v", err)
		}
		Grossamount := int(grossFloat)

		provider := ""
		if len(data.PaymentResult.VaNumbers) > 0 {
			provider = data.PaymentResult.VaNumbers[0].Bank
		}

		pembayaran := models.Pembayaran{
			KodeTransaksi:      data.PaymentResult.TransactionId,
			KodeOrderTransaksi: data.PaymentResult.OrderId,
			Provider:           provider,
			Amount:             int32(Grossamount),
			PaymentType:        data.PaymentResult.PaymentType,
			PaidAt:             data.PaymentResult.TransactionTime,
		}
		if err := tx.Create(&pembayaran).Error; err != nil {
			return err
		}

		for _, keranjang := range data.DataHold {
			var pembayaranObj models.Pembayaran
			if err := tx.Where(&models.Pembayaran{
				KodeTransaksi:      data.PaymentResult.TransactionId,
				KodeOrderTransaksi: data.PaymentResult.OrderId,
				Provider:           provider,
				Amount:             int32(Grossamount),
				PaymentType:        data.PaymentResult.PaymentType,
				PaidAt:             data.PaymentResult.TransactionTime,
			}).First(&pembayaranObj).Error; err != nil {
				return err
			}

			if pembayaranObj.ID == 0 {
				return fmt.Errorf("gagal, kredensial pembayaran tidak valid")
			}

			transaksi := models.Transaksi{
				IdPengguna:    keranjang.IDUser,
				IdSeller:      keranjang.IDSeller,
				IdBarangInduk: keranjang.IdBarangInduk,
				IdAlamat:      data.IdAlamatUser,
				IdPembayaran:  pembayaranObj.ID,
				KodeOrder:     data.PaymentResult.OrderId,
				Status:        "Dibayar",
				Metode:        data.PaymentResult.PaymentType,
				Kuantitas:     int16(keranjang.Dipesan),
				Total:         keranjang.HargaKategori * keranjang.Dipesan,
			}

			if err := tx.Create(&transaksi).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{
					IdBarangInduk: keranjang.IdBarangInduk,
					IdKategori:    keranjang.IdKategoriBarang,
					IdTransaksi:   0,
					Status:        "Dipesan",
					HoldBy:        keranjang.IDUser,
					HolderEntity:  "Pengguna",
				}).
				Updates(&models.VarianBarang{
					Status:      "Diproses",
					IdTransaksi: transaksi.ID,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		fmt.Printf("[FATAL] Transaction rollback | Err=%v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseLockTransaksi{
				Message: "Gaga",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseLockTransaksi{
			Message: "Berhasil",
		},
	}
}
