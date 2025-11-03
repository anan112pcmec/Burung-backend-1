package pengguna_transaction_services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"

	barang_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/barang"
	transaksi_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/transaksi"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	payment_gateaway "github.com/anan112pcmec/Burung-backend-1/app/payment"
	payment_gerai "github.com/anan112pcmec/Burung-backend-1/app/payment/gerai"
	payment_va "github.com/anan112pcmec/Burung-backend-1/app/payment/virtual_account"
	payment_wallet "github.com/anan112pcmec/Burung-backend-1/app/payment/wallet"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Critical
// ////////////////////////////////////////////////////////////////////////////////////

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Checkout Barang User
// Befungsi Untuk membuat checkout Barang Sebelum Akhirnya melakukan transaksi
// ////////////////////////////////////////////////////////////////////////////////////

func CheckoutBarangUser(data PayloadCheckoutBarang, db *gorm.DB) *response.ResponseForm {
	services := "CheckoutBarangUser"
	log.Printf("[%s] Memulai proses checkout untuk user ID: %v", services, data.IdentitasPengguna.ID)

	// Validasi pengguna
	if _, status := data.IdentitasPengguna.Validating(db); !status {
		log.Printf("[%s] Kredensial pengguna tidak valid untuk user ID: %v", services, data.IdentitasPengguna.ID)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_transaction_pengguna.ResponseDataCheckout{
				Message: "Kredensial pengguna tidak valid.",
			},
		}
	}

	// Pastikan semua barang dari seller yang sama
	var firstSellerID int64 = 0
	for i, keranjang := range data.DataCheckout {
		if i == 0 {
			firstSellerID = int64(keranjang.IdSeller)
			log.Printf("[%s] Seller pertama terdeteksi: %v", services, firstSellerID)
		} else if keranjang.IdSeller != int32(firstSellerID) {
			log.Printf("[%s] Ditemukan perbedaan seller ID pada item %v: %v != %v", services, i, keranjang.IdSeller, firstSellerID)
			return &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: services,
				Payload: response_transaction_pengguna.ResponseDataCheckout{
					Message: "Semua barang harus dari seller yang sama.",
				},
			}
		}
	}

	var responseData []response_transaction_pengguna.CheckoutData
	var processCheckout []models.Keranjang

	// Proses transaksi checkout
	err := db.Transaction(func(tx *gorm.DB) error {
		log.Printf("[%s] Mulai transaksi checkout (%d item)...", services, len(data.DataCheckout))

		for idx, keranjang := range data.DataCheckout {
			log.Printf("[%s] [Item-%d] Proses checkout BarangIndukID=%v KategoriID=%v Count=%v",
				services, idx, keranjang.IdBarangInduk, keranjang.IdKategori, keranjang.Count)

			// Hitung stok
			var jumlahStok int64
			if err := tx.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{
					IdBarangInduk: keranjang.IdBarangInduk,
					IdKategori:    keranjang.IdKategori,
					Status:        barang_enums.Ready,
				}).Count(&jumlahStok).Error; err != nil {
				log.Printf("[%s] [Item-%d] Gagal menghitung stok: %v", services, idx, err)
				return err
			}
			log.Printf("[%s] [Item-%d] Jumlah stok ready: %v", services, idx, jumlahStok)

			// Ambil detail barang induk
			var barang models.BarangInduk
			if err := tx.Select("nama_barang", "id_seller", "jenis_barang").
				Where(&models.BarangInduk{ID: keranjang.IdBarangInduk}).
				First(&barang).Error; err != nil {
				log.Printf("[%s] [Item-%d] Gagal ambil BarangInduk: %v", services, idx, err)
				return err
			}

			// Ambil kategori
			var kategori models.KategoriBarang
			if err := tx.Model(&models.KategoriBarang{}).Select("nama", "harga", "stok").
				Where(&models.KategoriBarang{ID: keranjang.IdKategori}).
				First(&kategori).Error; err != nil {
				log.Printf("[%s] [Item-%d] Gagal ambil KategoriBarang: %v", services, idx, err)
				return err
			}

			// Ambil nama seller
			var nama_seller string
			if errNamaSeller := tx.Model(&models.Seller{}).
				Select("nama").
				Where(&models.Seller{ID: barang.SellerID}).
				First(&nama_seller).Error; errNamaSeller != nil {
				log.Printf("[%s] [Item-%d] Gagal ambil Nama Seller: %v", services, idx, errNamaSeller)
				return errNamaSeller
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

			// Cek stok cukup
			if jumlahStok >= int64(keranjang.Count) {
				var varianIDs []int64
				if err := tx.Model(&models.VarianBarang{}).
					Where(&models.VarianBarang{
						IdBarangInduk: kategori.IdBarangInduk,
						IdKategori:    keranjang.IdKategori,
						Status:        barang_enums.Ready,
					}).
					Limit(int(keranjang.Count)).
					Pluck("id", &varianIDs).Error; err != nil {
					log.Printf("[%s] [Item-%d] Gagal ambil varian ID: %v", services, idx, err)
					resp.Message = "Terjadi kesalahan pada server. Silakan coba lagi nanti."
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				}

				if len(varianIDs) < int(keranjang.Count) {
					shortfall := int64(keranjang.Count) - int64(len(varianIDs))
					resp.Message = fmt.Sprintf("Stok kurang %v barang.", shortfall)
					resp.Status = false
					responseData = append(responseData, resp)
					log.Printf("[%s] [Item-%d] Stok kurang %v barang", services, idx, shortfall)
					continue
				}

				// Update status varian
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Updates(&models.VarianBarang{
						Status:       barang_enums.Dipesan,
						HoldBy:       data.IdentitasPengguna.ID,
						HolderEntity: "Pengguna",
					}).Error; err != nil {
					log.Printf("[%s] [Item-%d] Gagal update varian jadi Dipesan: %v", services, idx, err)
					resp.Message = "Terjadi kesalahan pada server. Silakan coba lagi nanti."
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				} else {
					log.Printf("[%s] [Item-%d] Varian ID terupdate: %+v", services, idx, varianIDs)
					var stokSaatIni int64 = 0
					_ = tx.Model(&models.KategoriBarang{}).Select("stok").
						Where(&models.KategoriBarang{ID: keranjang.IdKategori}).
						Take(&stokSaatIni)
					_ = tx.Model(&models.KategoriBarang{}).
						Where(&models.KategoriBarang{ID: keranjang.IdKategori}).
						Updates(map[string]interface{}{
							"stok": int32(stokSaatIni) - int32(keranjang.Count),
						})
					log.Printf("[%s] [Item-%d] Stok kategori dikurangi jadi %v", services, idx, int32(stokSaatIni)-int32(keranjang.Count))
				}

				resp.Message = "Barang siap untuk transaksi."
				resp.Status = true
				log.Printf("[%s] [Item-%d] Barang siap transaksi (stok mencukupi)", services, idx)
			} else {
				shortfall := int64(keranjang.Count) - jumlahStok
				resp.Message = fmt.Sprintf("Stok kurang %v barang.", shortfall)
				resp.Status = false
				log.Printf("[%s] [Item-%d] Stok tidak mencukupi, kurang %v", services, idx, shortfall)
			}

			processCheckout = append(processCheckout, keranjang)
			responseData = append(responseData, resp)
		}

		log.Printf("[%s] Transaksi selesai, total item diproses: %d", services, len(processCheckout))
		return nil
	})

	if err != nil {
		log.Printf("[%s] Gagal menjalankan transaksi checkout: %v", services, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseDataCheckout{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	log.Printf("[%s] Checkout selesai untuk user ID: %v dengan %d item", services, data.IdentitasPengguna.ID, len(responseData))

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseDataCheckout{
			Message:      "Checkout berhasil.",
			DataResponse: responseData,
			LayananPengiriman: response_transaction_pengguna.LayananPengiriman{
				JenisLayananKurir: data.JenisLayananKurir,
			},
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Batal Checkout User
// Befungsi Untuk MembatalkanCheckout yang telah dilakukan
// ////////////////////////////////////////////////////////////////////////////////////

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
				Where(models.VarianBarang{IdBarangInduk: keranjang.IdBarangInduk, IdKategori: keranjang.IdKategoriBarang, Status: barang_enums.Dipesan, HoldBy: keranjang.IDUser}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varianIDs).Error; err != nil {
				resp.Message = "Terjadi kesalahan pada server. Silakan coba lagi nanti."
				resp.Status = false
				responseData = append(responseData, resp)
				return err
			}

			if len(varianIDs) > 0 {
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Updates(map[string]interface{}{
						"status":        barang_enums.Ready,
						"hold_by":       0,
						"holder_entity": "",
					}).Error; err != nil {
					resp.Message = "Terjadi kesalahan pada server. Silakan coba lagi nanti."
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				} else {
					var stok_saat_ini int64 = 0
					_ = tx.Model(&models.KategoriBarang{}).Select("stok").Where(&models.KategoriBarang{
						ID: keranjang.IdKategoriBarang,
					}).Take(&stok_saat_ini)

					_ = tx.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
						ID: keranjang.IdKategoriBarang,
					}).Updates(&models.KategoriBarang{
						Stok: keranjang.Dipesan + int32(stok_saat_ini),
					})
				}

				resp.Message = "Checkout berhasil dibatalkan."
				resp.Status = true
			} else {
				resp.Message = "Tidak ada barang dipesan untuk dibatalkan."
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
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseDataCheckout{
			Message:      "Berhasil membatalkan checkout.",
			DataResponse: responseData,
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Formatting Transaksi
// Befungsi Untuk membantu snap trnsaksi dalam memformatkan sebuah transaksi
// ////////////////////////////////////////////////////////////////////////////////////

func FormattingTransaksi(user models.Pengguna, alamat models.AlamatPengguna, data response_transaction_pengguna.ResponseDataCheckout, db *gorm.DB, PaymentMethod, jenis_layanan string) (*response.ResponseForm, *snap.Request) {
	services := "ValidateTransaksi"
	fmt.Println("[TRACE] Start FormattingTransaksi")

	fmt.Println("[TRACE] Generate PaymentCode")
	var PaymentCode string
	var err_payment error
	maxRetry := 10
	for i := 0; i < maxRetry; i++ {
		PaymentCode, err_payment = helper.GenerateAutoPaymentId(db)
		if err_payment == nil {
			fmt.Printf("[TRACE] PaymentCode berhasil dibuat: %s (percobaan ke-%d)\n", PaymentCode, i+1)
			break
		} else {
			fmt.Printf("[TRACE] Gagal generate PaymentCode (percobaan ke-%d): %v\n", i+1, err_payment)
		}
	}

	if err_payment != nil {
		fmt.Println("[TRACE] Error fatal: gagal generate PaymentCode setelah retry")
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}, nil
	}

	fmt.Println("[TRACE] Siapkan AlamatPengguna")
	AlamatPengguna := midtrans.CustomerAddress{
		Address:     alamat.NamaAlamat,
		City:        alamat.Kota,
		Postcode:    alamat.KodePos,
		CountryCode: alamat.KodeNegara,
	}

	fmt.Println("[TRACE] Generate ItemDetail dan TotalHarga")
	items, TotalHarga := helper.GenerateItemDetail(data, db, jenis_layanan)
	fmt.Printf("[TRACE] TotalHarga: %v\n", TotalHarga)

	var PM []snap.SnapPaymentType
	fmt.Printf("[TRACE] PaymentMethod: %s\n", PaymentMethod)

	switch PaymentMethod {
	case "va":
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeBCAVA,
			snap.PaymentTypeBNIVA,
			snap.PaymentTypeBRIVA,
			snap.PaymentTypePermataVA,
		}
	case "wallet":
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeGopay,
			snap.PaymentTypeShopeepay,
		}
	case "gerai":
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeIndomaret,
			snap.PaymentTypeAlfamart,
		}
	case "credit":
		PM = []snap.SnapPaymentType{
			snap.PaymentTypeAkulaku,
			snap.PaymentTypeCreditCard,
		}
	default:
		fmt.Println("[TRACE] PaymentMethod tidak dikenali, daftar kosong")
	}

	fmt.Println("[TRACE] Buat SnapRequest")
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

	fmt.Println("[TRACE] Selesai membuat SnapRequest, return response")
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}, SnapReqeust
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Validating Transaksi
// Befungsi Untuk membantu snap transaksi untuk membuat sebuah transaksi
// ////////////////////////////////////////////////////////////////////////////////////

func ValidateTransaksi(snapReq *snap.Request) (*snap.Response, *response.ResponseForm) {
	services := "ProsesTransaksiDenganSDK"
	fmt.Println("[TRACE] Start ValidateTransaksi")

	if snapReq == nil {
		fmt.Println("[TRACE] snapReq nil — request tidak valid")
		return nil, &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
		}
	}

	fmt.Println("[TRACE] Inisialisasi snap.Client")
	var s snap.Client
	s.New("Mid-server-7wpABbBW_WURdLxcxc5bX5eb", midtrans.Sandbox)

	fmt.Println("[TRACE] Membuat transaksi dengan Snap SDK")
	snapResp, err := s.CreateTransaction(snapReq)
	if err != nil {
		fmt.Printf("[TRACE] Gagal membuat transaksi: %v\n", err)
		return nil, &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	fmt.Println("[TRACE] Transaksi berhasil dibuat via Snap SDK")
	if snapResp != nil {
		fmt.Printf("[TRACE] OrderID: %s\n", snapResp.Token)
		fmt.Printf("[TRACE] RedirectURL: %s\n", snapResp.RedirectURL)
	} else {
		fmt.Println("[TRACE] Warning: snapResp nil meskipun tidak ada error")
	}

	fmt.Println("[TRACE] Selesai ValidateTransaksi, return response")
	return snapResp, &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Snap Transaksi
// Fungsi yang melayani api pada pengguna dan memanfaaykan Validate Transaksi Dan Formatting transaksi(2 fungsi
// pendukungnya)
// ////////////////////////////////////////////////////////////////////////////////////

func SnapTransaksi(data PayloadSnapTransaksiRequest, db *gorm.DB) *response.ResponseForm {
	services := "SnapTransaksiUser"
	fmt.Println("[TRACE] Start SnapTransaksi")

	model, status := data.IdentitasPengguna.Validating(db)
	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal Validasi User Tidak Valid",
		}
	}

	for _, dc := range data.DataCheckout.DataResponse {
		var errcheck bool = false
		var hitung int64 = 0
		if err := db.Model(&models.VarianBarang{}).Where(&models.VarianBarang{
			IdBarangInduk: dc.IdBarangInduk,
			IdKategori:    dc.IdKategoriBarang,
			Status:        barang_enums.Dipesan,
			HoldBy:        dc.IDUser,
		}).Count(&hitung).Error; err != nil {
			errcheck = true
		}

		if hitung != int64(dc.Dipesan) {
			errcheck = true
		}

		if errcheck {
			return &response.ResponseForm{
				Status:   http.StatusUnavailableForLegalReasons,
				Services: services,
			}
		}
	}

	if data.AlamatInformation.NamaAlamat == "" && data.AlamatInformation.KodeNegara == "" && data.AlamatInformation.IDPengguna != data.AlamatInformation.ID && data.AlamatInformation.NomorTelephone == "" {
		fmt.Println("[TRACE] Data alamat tidak valid atau tidak sesuai dengan pengguna")
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	fmt.Println("[TRACE] Memulai FormattingTransaksi()")
	SnapErr, SnapReq := FormattingTransaksi(
		model,
		data.AlamatInformation,
		data.DataCheckout,
		db,
		data.PaymentMethod,
		data.DataCheckout.LayananPengiriman.JenisLayananKurir,
	)

	if SnapErr.Status != http.StatusOK {
		fmt.Printf("[TRACE] FormattingTransaksi gagal (status: %d)\n", SnapErr.Status)
		return &response.ResponseForm{
			Status:   SnapErr.Status,
			Services: services,
		}
	}
	fmt.Println("[TRACE] FormattingTransaksi sukses, lanjut ke ValidateTransaksi()")

	SnapResponse, SnapResponseErr := ValidateTransaksi(SnapReq)
	if SnapResponseErr.Status != http.StatusOK {
		fmt.Printf("[TRACE] ValidateTransaksi gagal (status: %d)\n", SnapResponseErr.Status)
		return &response.ResponseForm{
			Status:   SnapErr.Status,
			Services: services,
		}
	}

	fmt.Println("[TRACE] ValidateTransaksi sukses, siapkan payload response akhir")

	if SnapResponse != nil {
		fmt.Printf("[TRACE] SnapResponse Token: %s\n", SnapResponse.Token)
		fmt.Printf("[TRACE] SnapResponse RedirectURL: %s\n", SnapResponse.RedirectURL)
	} else {
		fmt.Println("[TRACE] Warning: SnapResponse nil meskipun tidak error")
	}

	fmt.Println("[TRACE] Selesai SnapTransaksi, return response ke client")
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

// func PendingTransaksi(ctx context.Context, data PayloadPendingTransaksi, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
// 	services := "PendingTransaksi"

// 	if _, status := data.IdentitasPengguna.Validating(db); !status {
// 		return &response.ResponseForm{
// 			Status:   http.StatusNotFound,
// 			Services: services,
// 			Payload: response_transaction_pengguna.ResponsePendingTransaksi{
// 				Message: "Gagal Kredensial User Tidak Valid",
// 			},
// 		}
// 	}

// 	key := fmt.Sprintf("transaction_pengguna_pending_id:%v:transaction_code:%s",
// 		data.IdentitasPengguna.ID, data.DataPending.OrderId)

// 	fields := map[string]interface{}{
// 		"finish_redirect_url": data.DataPending.FinishRedirectUrl,
// 		"fraud_status":        data.DataPending.FraudStatus,
// 		"gross_amount":        data.DataPending.GrossAmout,
// 		"order_id":            data.DataPending.OrderId,
// 		"payment_type":        data.DataPending.PaymentType,
// 		"status_code":         data.DataPending.StatusCode,
// 		"status_message":      data.DataPending.StatusMessage,
// 		"transaction_id":      data.DataPending.TransactionId,
// 		"transaction_status":  data.DataPending.TransactionStatus,
// 		"transaction_time":    data.DataPending.TransactionTime,
// 	}

// 	if err := rds.HSet(ctx, key, fields).Err(); err != nil {
// 		fmt.Println("⚠️ Gagal menyimpan ke Redis:", err)
// 		return &response.ResponseForm{
// 			Status:   http.StatusInternalServerError,
// 			Services: services,
// 			Payload: response_transaction_pengguna.ResponsePendingTransaksi{
// 				Message: "Gagal, Server sedang sibuk coba lagi lain waktu",
// 			},
// 		}
// 	}

// 	return &response.ResponseForm{
// 		Status:   http.StatusOK,
// 		Services: services,
// 		Payload: response_transaction_pengguna.ResponsePendingTransaksi{
// 			Message: "Berhasil",
// 		},
// 	}
// }

// func CallPendingTransaksi(data PayloadCallPendingTransaksi, rds *redis.Client) *response.ResponseForm {
// 	services := "CallPendingTransaksi"

//		return &response.ResponseForm{
//			Status:   http.StatusOK,
//			Services: services,
//		}
//	}
//
// ***** INFO ******
// Skema Pending Akan Tersedia di saat mendatang untuk saat ini semua transaksi yang tak sengaja di bayar tak masuk pending melainkan akan langsung di batalkan

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Batal Transaksi
// Befungsi Untuk Membatalkan Transaksi Yang Telah Dibuat SnapTransaksi lewat Validating Transaksi
// Semua yang telah melibatkan payment gateaway yang kemudian tidak melanjutkan pembayaran akan di batalkan
// oleh fungsi ini
// ////////////////////////////////////////////////////////////////////////////////////

func BatalTransaksi(data response_transaction_pengguna.SnapTransaksi, db *gorm.DB) *response.ResponseForm {
	services := "BatalTransaksi"

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, keranjang := range data.DataCheckout {

			var varianIDs []int64

			if err := tx.Model(&models.VarianBarang{}).
				Where(models.VarianBarang{IdBarangInduk: keranjang.IdBarangInduk, IdKategori: keranjang.IdKategoriBarang, Status: barang_enums.Dipesan, HoldBy: keranjang.IDUser}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varianIDs).Error; err != nil {
				return err
			}

			if len(varianIDs) > 0 {
				if err := tx.Model(&models.VarianBarang{}).
					Where("id IN ?", varianIDs).
					Updates(map[string]interface{}{
						"status":        barang_enums.Ready,
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
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseBatalTransaksi{
			Message: "Transaksi berhasil dibatalkan.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Simpan Transaksi
// Befungsi Untuk Menyimpan Data Transaksi yang nantinya fungsi ini akan di paggil di fungsi LockTransaksi
// ////////////////////////////////////////////////////////////////////////////////////

func SimpanTransaksi(pembayaran *models.Pembayaran, DataHold *[]response_transaction_pengguna.CheckoutData, IdAlamatUser int64, tx *gorm.DB) error {
	fmt.Println("=== [TRACE] Mulai SimpanTransaksi ===")
	fmt.Printf("Pembayaran Info: %+v\n", pembayaran)
	fmt.Printf("Jumlah DataHold: %d\n", len(*DataHold))
	fmt.Printf("IdAlamatUser: %d\n", IdAlamatUser)

	for i, keranjang := range *DataHold {
		fmt.Printf("\n[TRACE] Memproses keranjang ke-%d\n", i+1)
		fmt.Printf("Data Keranjang: %+v\n", keranjang)

		var pembayaranObj models.Pembayaran
		fmt.Println("[TRACE] Mengecek kredensial pembayaran di database...")

		if err := tx.Where(&models.Pembayaran{
			KodeTransaksi:      pembayaran.KodeTransaksi,
			KodeOrderTransaksi: pembayaran.KodeOrderTransaksi,
			Provider:           pembayaran.Provider,
			Amount:             pembayaran.Amount,
			PaymentType:        pembayaran.PaymentType,
			PaidAt:             pembayaran.PaidAt,
		}).First(&pembayaranObj).Error; err != nil {
			fmt.Printf("[ERROR] Gagal menemukan pembayaran: %v\n", err)
			return fmt.Errorf("gagal mencari pembayaran di database: %w", err)
		}

		fmt.Printf("[TRACE] Pembayaran ditemukan: ID=%d, KodeOrder=%s\n", pembayaranObj.ID, pembayaranObj.KodeOrderTransaksi)

		if pembayaranObj.ID == 0 {
			fmt.Println("[ERROR] Kredensial pembayaran tidak valid (ID=0)")
			return fmt.Errorf("kredensial pembayaran tidak valid")
		}

		transaksi := models.Transaksi{
			IdPengguna:       keranjang.IDUser,
			IdSeller:         keranjang.IDSeller,
			IdBarangInduk:    keranjang.IdBarangInduk,
			IdKategoriBarang: keranjang.IdKategoriBarang,
			IdAlamat:         IdAlamatUser,
			IdPembayaran:     pembayaranObj.ID,
			KodeOrder:        pembayaranObj.KodeOrderTransaksi,
			Status:           transaksi_enums.Dibayar,
			Metode:           pembayaranObj.PaymentType,
			Kuantitas:        int16(keranjang.Dipesan),
			Total:            keranjang.HargaKategori * keranjang.Dipesan,
		}

		fmt.Printf("[TRACE] Membuat transaksi baru: %+v\n", transaksi)

		if err := tx.Create(&transaksi).Error; err != nil {
			fmt.Printf("[ERROR] Gagal membuat transaksi: %v\n", err)
			return fmt.Errorf("gagal membuat transaksi: %w", err)
		}
		fmt.Printf("[TRACE] Transaksi berhasil dibuat dengan ID=%d\n", transaksi.ID)

		fmt.Println("[TRACE] Mengupdate varian barang terkait menjadi status 'Diproses'...")

		if err := tx.Model(&models.VarianBarang{}).
			Where(&models.VarianBarang{
				IdBarangInduk: keranjang.IdBarangInduk,
				IdKategori:    keranjang.IdKategoriBarang,
				IdTransaksi:   0,
				Status:        barang_enums.Dipesan,
				HoldBy:        keranjang.IDUser,
				HolderEntity:  "Pengguna",
			}).
			Updates(&models.VarianBarang{
				Status:      barang_enums.Diproses,
				IdTransaksi: transaksi.ID,
			}).Error; err != nil {
			fmt.Printf("[ERROR] Gagal update varian barang: %v\n", err)
			return fmt.Errorf("gagal update varian barang: %w", err)
		}

		fmt.Println("[TRACE] Varian barang berhasil diperbarui menjadi 'Diproses'")
	}

	fmt.Println("\n=== [TRACE] Selesai SimpanTransaksi tanpa error ===")
	return nil
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Lock Transaksi VA
// Befungsi saat sebuah transaksi sudah di bayar, setelah transaksi di bayar maka fungsi
// lock transaksi akan menjalankan rentetan yang perlu di jalankan ke db utama sesuai dengan
// jenis pembayaran yang dilakukan oleh pengguna disini adalah VA (virtual account)
// ////////////////////////////////////////////////////////////////////////////////////

func LockTransaksiVa(data PayloadLockTransaksiVa, db *gorm.DB) *response.ResponseForm {
	services := "LockTransaksiVa"

	for _, keranjang := range data.DataHold {
		if keranjang.IDSeller == 0 && keranjang.IDUser == 0 && keranjang.IdBarangInduk == 0 {
			return &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: services,
				Payload: response_transaction_pengguna.ResponseLockTransaksi{
					Message: "Data keranjang tidak valid.",
				},
			}
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		bank, err_p := payment_gateaway.ParseVirtualAccount(data.PaymentResult)
		if err_p != nil {
			return err_p
		}

		var (
			resp payment_va.Response
		)

		d, err_m := json.Marshal(data.PaymentResult)
		if err_m != nil {
			return err_m
		}

		switch bank {
		case "bca":
			var obj payment_va.BcaVirtualAccountResponse
			if err := json.Unmarshal(d, &obj); err != nil {
				return err
			}
			resp = &obj

		case "bni":
			var obj payment_va.BniVirtualAccountResponse
			if err := json.Unmarshal(d, &obj); err != nil {
				return err
			}
			resp = &obj

		case "bri":
			var obj payment_va.BriVirtualAccountResponse
			if err := json.Unmarshal(d, &obj); err != nil {
				return err
			}
			resp = &obj

		case "permata":
			var obj payment_va.PermataVirtualAccount
			if err := json.Unmarshal(d, &obj); err != nil {
				return err
			}
			resp = &obj

		default:
			return fmt.Errorf("bank tidak dikenali: %v", bank)
		}

		pembayaran, ok := resp.Pembayaran()
		if !ok {
			return fmt.Errorf("gagal membuat pembayaran %s", bank)
		}
		if err := tx.Create(&pembayaran).Error; err != nil {
			return err
		}

		//
		// Sanitasi Id Pengguna
		//
		pembayaran.IdPengguna = data.DataHold[0].IDUser

		status := SimpanTransaksi(&pembayaran, &data.DataHold, data.IdAlamatUser, tx)

		return status
	}); err != nil {
		fmt.Printf("[ERROR] Transaction rollback | Err=%v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseLockTransaksi{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseLockTransaksi{
			Message: "Transaksi berhasil dikunci.",
		},
	}
}

func PaidFailedTransaksiVa(data PayloadPaidFailedTransaksiVa, db *gorm.DB) *response.ResponseForm {
	services := "PaidFailedTransaksiVa"

	// --- Parse bank VA ---
	bank, err_p := payment_gateaway.ParseVirtualAccount(data.PaymentResult)
	if err_p != nil {
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			},
		}
	}

	// --- Encode/Decode payment result ---
	raw, err_m := json.Marshal(data.PaymentResult)
	if err_m != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			},
		}
	}

	var resp payment_va.Response

	switch bank {
	case "bca":
		var obj payment_va.BcaVirtualAccountResponse
		if err := json.Unmarshal(raw, &obj); err != nil {
			return &response.ResponseForm{Status: http.StatusBadRequest, Services: services, Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			}}
		}
		resp = &obj

	case "bni":
		var obj payment_va.BniVirtualAccountResponse
		if err := json.Unmarshal(raw, &obj); err != nil {
			return &response.ResponseForm{Status: http.StatusBadRequest, Services: services, Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			}}
		}
		resp = &obj

	case "bri":
		var obj payment_va.BriVirtualAccountResponse
		if err := json.Unmarshal(raw, &obj); err != nil {
			return &response.ResponseForm{Status: http.StatusBadRequest, Services: services, Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			}}
		}
		resp = &obj

	case "permata":
		var obj payment_va.PermataVirtualAccount
		if err := json.Unmarshal(raw, &obj); err != nil {
			return &response.ResponseForm{Status: http.StatusBadRequest, Services: services, Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			}}
		}
		resp = &obj

	default:
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Bank tidak dikenali",
			},
		}
	}

	if resp == nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengenali virtual account",
			},
		}
	}

	standard_response, ok := resp.StandardResponse()
	if !ok {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: "Gagal mengambil standard response",
			},
		}
	}

	standard_response.IdPengguna = data.DataHold[0].IDUser

	// --- Jalankan transaksi database ---
	err := db.Transaction(func(tx *gorm.DB) error {
		// Simpan ke PaidFailed
		if err := tx.Create(&standard_response).Error; err != nil {
			return fmt.Errorf("gagal menyimpan PaidFailed: %w", err)
		}

		// Ambil ID PaidFailed
		var pf int64 = 0
		if err := tx.Model(&models.PembayaranFailed{}).
			Select("id").Where(&models.PembayaranFailed{
			IdPengguna:    data.DataHold[0].IDUser,
			OrderId:       standard_response.OrderId,
			TransactionId: standard_response.TransactionId,
		}).Limit(1).Take(&pf).Error; err != nil {
			return fmt.Errorf("gagal mengambil id PaidFailed: %w", err)
		}

		if pf == 0 {
			return fmt.Errorf("id PaidFailed tidak ditemukan")
		}

		// Simpan TransaksiFailed per item
		for i, d := range data.DataHold {
			tf := models.TransaksiFailed{
				IdPembayaranFailed: pf,
				IdPengguna:         d.IDUser,
				IdSeller:           d.IDSeller,
				IdBarangInduk:      d.IdBarangInduk,
				IdKategoriBarang:   d.IdKategoriBarang,
				IdAlamat:           data.IdAlamatUser,
				Catatan:            d.Message,
				Kuantitas:          int16(d.Dipesan),
				Total:              int64(d.Dipesan) * int64(d.HargaKategori),
				JenisPengiriman:    data.JenisLayananKurir,
			}

			if err := tx.Create(&tf).Error; err != nil {
				return fmt.Errorf("gagal menyimpan transaksi ke-%d: %w", i+1, err)
			}
		}

		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: fmt.Sprintf("Transaksi gagal: %v", err),
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
			Message: "Berhasil",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Lock Transaksi Wallet
// Befungsi saat sebuah transaksi sudah di bayar, setelah transaksi di bayar maka fungsi
// lock transaksi akan menjalankan rentetan yang perlu di jalankan ke db utama sesuai dengan
// jenis pembayaran yang dilakukan oleh pengguna disini adalah Wallet
// ////////////////////////////////////////////////////////////////////////////////////

func LockTransaksiWallet(data PayloadLockTransaksiWallet, db *gorm.DB) *response.ResponseForm {
	services := "LockTransaksiWallet"

	for _, keranjang := range data.DataHold {
		if keranjang.IDSeller == 0 && keranjang.IDUser == 0 && keranjang.IdBarangInduk == 0 {
			return &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: services,
				Payload: response_transaction_pengguna.ResponseLockTransaksi{
					Message: "Data keranjang tidak valid.",
				},
			}
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {

		pembayaran, ok := data.PaymentResult.Pembayaran()
		if !ok {
			return fmt.Errorf("gagal memproses pembayaran wallet")
		}

		//
		// Sanitasi Id Pengguna
		//
		pembayaran.IdPengguna = data.DataHold[0].IDUser

		if err := tx.Create(&pembayaran).Error; err != nil {
			return err
		}

		status := SimpanTransaksi(&pembayaran, &data.DataHold, data.IdAlamatUser, tx)

		return status
	}); err != nil {
		fmt.Printf("[ERROR] Transaction rollback | Err=%v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseLockTransaksi{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseLockTransaksi{
			Message: "Transaksi berhasil dikunci.",
		},
	}
}

func PaidFailedTransaksiWallet(data PayloadPaidFailedTransaksiWallet, db *gorm.DB) *response.ResponseForm {
	services := "PaidFailedTransaksiWallet"

	var resp payment_wallet.Response = &data.PaymentResult
	standard_response, _ := resp.StandardResponse()

	standard_response.IdPengguna = data.DataHold[0].IDUser
	// --- Jalankan transaksi database ---
	err := db.Transaction(func(tx *gorm.DB) error {
		// Simpan ke PaidFailed
		if err := tx.Create(&standard_response).Error; err != nil {
			return fmt.Errorf("gagal menyimpan PaidFailed: %w", err)
		}

		// Ambil ID PaidFailed
		var pf int64 = 0
		if err := tx.Model(&models.PembayaranFailed{}).
			Select("id").Where(&models.PembayaranFailed{
			IdPengguna:    data.DataHold[0].IDUser,
			OrderId:       standard_response.OrderId,
			TransactionId: standard_response.TransactionId,
		}).Limit(1).Take(&pf).Error; err != nil {
			return fmt.Errorf("gagal mengambil id PaidFailed: %w", err)
		}

		if pf == 0 {
			return fmt.Errorf("id PaidFailed tidak ditemukan")
		}

		// Simpan TransaksiFailed per item
		for i, d := range data.DataHold {
			tf := models.TransaksiFailed{
				IdPembayaranFailed: pf,
				IdPengguna:         d.IDUser,
				IdSeller:           d.IDSeller,
				IdBarangInduk:      d.IdBarangInduk,
				IdKategoriBarang:   d.IdKategoriBarang,
				IdAlamat:           data.IdAlamatUser,
				Catatan:            d.Message,
				Kuantitas:          int16(d.Dipesan),
				Total:              int64(d.Dipesan) * int64(d.HargaKategori),
			}

			if err := tx.Create(&tf).Error; err != nil {
				return fmt.Errorf("gagal menyimpan transaksi ke-%d: %w", i+1, err)
			}
		}

		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: fmt.Sprintf("Transaksi gagal: %v", err),
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
			Message: "Berhasil",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Lock Transaksi Gerai
// Befungsi saat sebuah transaksi sudah di bayar, setelah transaksi di bayar maka fungsi
// lock transaksi akan menjalankan rentetan yang perlu di jalankan ke db utama sesuai dengan
// jenis pembayaran yang dilakukan oleh pengguna disini adalah Gerai
// ////////////////////////////////////////////////////////////////////////////////////

func LockTransaksiGerai(data PayloadLockTransaksiGerai, db *gorm.DB) *response.ResponseForm {
	services := "LockTransaksiGerai"

	for _, keranjang := range data.DataHold {
		if keranjang.IDSeller == 0 && keranjang.IDUser == 0 && keranjang.IdBarangInduk == 0 {
			return &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: services,
				Payload: response_transaction_pengguna.ResponseLockTransaksi{
					Message: "Data keranjang tidak valid.",
				},
			}
		}
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		var (
			resp payment_gerai.Response
		)
		resp = &data.PaymentResult

		pembayaran, ok := resp.Pembayaran()
		if !ok {
			return fmt.Errorf("gagal memproses pembayaran wallet")
		}

		//
		// Sanitasi Id Pengguna
		//
		pembayaran.IdPengguna = data.DataHold[0].IDUser

		if err := tx.Create(&pembayaran).Error; err != nil {
			return err
		}

		status := SimpanTransaksi(&pembayaran, &data.DataHold, data.IdAlamatUser, tx)

		return status
	}); err != nil {
		fmt.Printf("[ERROR] Transaction rollback | Err=%v\n", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseLockTransaksi{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponseLockTransaksi{
			Message: "Transaksi berhasil dikunci.",
		},
	}
}

func PaidFailedTransaksiGerai(data PayloadPaidFailedTransaksiGerai, db *gorm.DB) *response.ResponseForm {
	services := "PaidFailedTransaksiGerai"

	var resp payment_gerai.Response = &data.PaymentResult
	standard_response, _ := resp.StandardResponse()

	standard_response.IdPengguna = data.DataHold[0].IDUser

	// --- Jalankan transaksi database ---
	err := db.Transaction(func(tx *gorm.DB) error {
		// Simpan ke PaidFailed
		if err := tx.Create(&standard_response).Error; err != nil {
			return fmt.Errorf("gagal menyimpan PaidFailed: %w", err)
		}

		// Ambil ID PaidFailed
		var pf int64 = 0
		if err := tx.Model(&models.PembayaranFailed{}).
			Select("id").Where(&models.PembayaranFailed{
			IdPengguna:    data.DataHold[0].IDUser,
			OrderId:       standard_response.OrderId,
			TransactionId: standard_response.TransactionId,
		}).Limit(1).Take(&pf).Error; err != nil {
			return fmt.Errorf("gagal mengambil id PaidFailed: %w", err)
		}

		if pf == 0 {
			return fmt.Errorf("id PaidFailed tidak ditemukan")
		}

		// Simpan TransaksiFailed per item
		for i, d := range data.DataHold {
			tf := models.TransaksiFailed{
				IdPembayaranFailed: pf,
				IdPengguna:         d.IDUser,
				IdSeller:           d.IDSeller,
				IdBarangInduk:      d.IdBarangInduk,
				IdKategoriBarang:   d.IdKategoriBarang,
				IdAlamat:           data.IdAlamatUser,
				Catatan:            d.Message,
				Kuantitas:          int16(d.Dipesan),
				Total:              int64(d.Dipesan) * int64(d.HargaKategori),
			}

			if err := tx.Create(&tf).Error; err != nil {
				return fmt.Errorf("gagal menyimpan transaksi ke-%d: %w", i+1, err)
			}
		}

		return nil
	})

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
				Message: fmt.Sprintf("Transaksi gagal: %v", err),
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_transaction_pengguna.ResponsePaidFailedTransaksi{
			Message: "Berhasil",
		},
	}
}
