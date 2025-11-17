package pengguna_transaction_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"

	barang_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/barang"
	transaksi_enums "github.com/anan112pcmec/Burung-backend-1/app/database/enums/transaksi"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
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

func CheckoutBarangUser(ctx context.Context, data PayloadCheckoutBarang, db *gorm.DB) *response.ResponseForm {
	services := "CheckoutBarangUser"
	log.Printf("[%s] Memulai proses checkout untuk user ID: %v", services, data.IdentitasPengguna.ID)

	// Validasi pengguna
	if _, status := data.IdentitasPengguna.Validating(ctx, db); !status {
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
	var firstSellerID int64
	for i, item := range data.DataCheckout {
		if i == 0 {
			firstSellerID = int64(item.IdSeller)
			continue
		}
		if item.IdSeller != int32(firstSellerID) {
			return &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: services,
				Payload: response_transaction_pengguna.ResponseDataCheckout{
					Message: "Semua barang harus dari seller yang sama.",
				},
			}
		}
	}

	responseData := make([]response_transaction_pengguna.CheckoutData, 0, len(data.DataCheckout))
	varianUpdates := make([]int64, 0)
	kategoriUpdates := make(map[int32]int32) // kategoriID => total jumlah dipesan

	err := db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(data.DataCheckout); i++ {
			// Hitung stok
			var jumlahStok int64
			if err := tx.Model(&models.VarianBarang{}).
				Where(&models.VarianBarang{
					IdBarangInduk: data.DataCheckout[i].IdBarangInduk,
					IdKategori:    data.DataCheckout[i].IdKategori,
					Status:        barang_enums.Ready,
				}).Count(&jumlahStok).Error; err != nil {
				return err
			}

			// Ambil detail barang induk
			var barang models.BarangInduk
			if err := tx.Select("nama_barang", "id_seller", "jenis_barang").
				Where(&models.BarangInduk{ID: data.DataCheckout[i].IdBarangInduk}).First(&barang).Error; err != nil {
				return err
			}

			// Ambil kategori
			var kategori models.KategoriBarang
			if err := tx.Select("nama", "harga", "stok", "id_barang_induk").
				Where(&models.KategoriBarang{ID: data.DataCheckout[i].IdKategori}).First(&kategori).Error; err != nil {
				return err
			}

			// Ambil nama seller
			var namaSeller string
			if err := tx.Model(&models.Seller{}).Select("nama").
				Where(&models.Seller{ID: barang.SellerID}).First(&namaSeller).Error; err != nil {
				return err
			}

			resp := response_transaction_pengguna.CheckoutData{
				IDUser:           data.IdentitasPengguna.ID,
				IDSeller:         data.DataCheckout[i].IdSeller,
				NamaSeller:       namaSeller,
				JenisBarang:      barang.JenisBarang,
				IdBarangInduk:    data.DataCheckout[i].IdBarangInduk,
				IdKategoriBarang: data.DataCheckout[i].IdKategori,
				HargaKategori:    kategori.Harga,
				NamaBarang:       barang.NamaBarang,
				NamaKategori:     kategori.Nama,
				Dipesan:          int32(data.DataCheckout[i].Jumlah),
			}

			if jumlahStok >= int64(data.DataCheckout[i].Jumlah) {
				var varianIDs []int64
				if err := tx.Model(&models.VarianBarang{}).
					Where(&models.VarianBarang{
						IdBarangInduk: kategori.IdBarangInduk,
						IdKategori:    data.DataCheckout[i].IdKategori,
						Status:        barang_enums.Ready,
					}).Limit(int(data.DataCheckout[i].Jumlah)).Pluck("id", &varianIDs).Error; err != nil {
					resp.Message = "Terjadi kesalahan pada server."
					resp.Status = false
					responseData = append(responseData, resp)
					return err
				}

				if len(varianIDs) < int(data.DataCheckout[i].Jumlah) {
					shortfall := int64(data.DataCheckout[i].Jumlah) - int64(len(varianIDs))
					resp.Message = fmt.Sprintf("Stok kurang %v barang.", shortfall)
					resp.Status = false
					responseData = append(responseData, resp)
					continue
				}

				varianUpdates = append(varianUpdates, varianIDs...)
				kategoriUpdates[int32(data.DataCheckout[i].IdKategori)] += int32(data.DataCheckout[i].Jumlah)

				resp.Message = "Barang siap untuk transaksi."
				resp.Status = true
			} else {
				shortfall := int64(data.DataCheckout[i].Jumlah) - jumlahStok
				resp.Message = fmt.Sprintf("Stok kurang %v barang.", shortfall)
				resp.Status = false
			}

			responseData = append(responseData, resp)
		}

		// Update status varian sekaligus
		if len(varianUpdates) > 0 {
			if err := tx.Model(&models.VarianBarang{}).
				Where("id IN ?", varianUpdates).
				Updates(&models.VarianBarang{
					Status:       barang_enums.Dipesan,
					HoldBy:       data.IdentitasPengguna.ID,
					HolderEntity: "Pengguna",
				}).Error; err != nil {
				return err
			}
		}

		// Update stok kategori secara atomic
		for kategoriID, totalDipesan := range kategoriUpdates {
			if err := tx.Model(&models.KategoriBarang{}).
				Where("id = ? AND stok >= ?", kategoriID, totalDipesan).
				UpdateColumn("stok", gorm.Expr("stok - ?", totalDipesan)).Error; err != nil {
				return err
			}
		}

		var idKeranjang []int64
		for i := 0; i < len(data.DataCheckout); i++ {
			idKeranjang = append(idKeranjang, data.DataCheckout[i].ID)
		}
		if err := db.WithContext(ctx).Model(&models.Keranjang{}).Where("id IN ?", idKeranjang).Delete(&models.Keranjang{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("[%s] Gagal checkout: %v", services, err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_transaction_pengguna.ResponseDataCheckout{
				Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			},
		}
	}

	// Hapus keranjang

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

	err := db.Transaction(func(tx *gorm.DB) error {
		var varianIDs []int64
		kategoriUpdates := make(map[int32]int32) // kategoriID => total jumlah dikembalikan

		for _, keranjang := range data.DataResponse {
			var varian_id []int64
			if err := tx.Model(&models.VarianBarang{}).
				Where(models.VarianBarang{
					IdBarangInduk: keranjang.IdBarangInduk,
					IdKategori:    keranjang.IdKategoriBarang,
					Status:        barang_enums.Dipesan,
					HoldBy:        keranjang.IDUser,
				}).
				Limit(int(keranjang.Dipesan)).
				Pluck("id", &varian_id).Error; err != nil {
				return err
			}
			varianIDs = append(varianIDs, varian_id...)
			kategoriUpdates[int32(keranjang.IdKategoriBarang)] += keranjang.Dipesan
		}

		// Update status semua varian sekaligus
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

		// Update stok kategori secara atomic
		for kategoriID, totalDikembalikan := range kategoriUpdates {
			if err := tx.Model(&models.KategoriBarang{}).
				Where("id = ?", kategoriID).
				UpdateColumn("stok", gorm.Expr("stok + ?", totalDikembalikan)).Error; err != nil {
				return err
			}
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
			Message: "Berhasil membatalkan checkout.",
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Snap Transaksi
// Fungsi yang melayani api pada pengguna dan memanfaaykan Validate Transaksi Dan Formatting transaksi(2 fungsi
// pendukungnya)
// ////////////////////////////////////////////////////////////////////////////////////

func SnapTransaksi(ctx context.Context, data PayloadSnapTransaksiRequest, db *gorm.DB) *response.ResponseForm {
	services := "SnapTransaksiUser"
	fmt.Println("[TRACE] Start SnapTransaksi")

	model, status := data.IdentitasPengguna.Validating(ctx, db)
	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal Validasi User Tidak Valid",
		}
	}

	var id_data_seller int64 = 0

	if err := db.WithContext(ctx).Model(&models.Seller{}).Select("id").Where(&models.Seller{
		ID: data.DataCheckout.DataResponse[0].IDSeller,
	}).Limit(1).Scan(&id_data_seller).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal seller tidak ditemukan",
		}
	}

	if id_data_seller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal Seller tidak valid",
		}
	}

	for i := 0; i < len(data.DataCheckout.DataResponse); i++ {
		var errcheck bool = false
		var varianIds []int64 = make([]int64, 0, int(data.DataCheckout.DataResponse[i].Dipesan))
		if err := db.WithContext(ctx).Model(&models.VarianBarang{}).Select("id").Where(&models.VarianBarang{
			IdBarangInduk: data.DataCheckout.DataResponse[i].IdBarangInduk,
			IdKategori:    data.DataCheckout.DataResponse[i].IdKategoriBarang,
			Status:        barang_enums.Dipesan,
			HoldBy:        data.DataCheckout.DataResponse[i].IDUser,
		}).Limit(int(data.DataCheckout.DataResponse[i].Dipesan)).Scan(&varianIds).Error; err != nil {
			errcheck = true
		}

		if len(varianIds) != int(data.DataCheckout.DataResponse[i].Dipesan) {
			errcheck = true
		}

		if errcheck {
			return &response.ResponseForm{
				Status:   http.StatusUnavailableForLegalReasons,
				Services: services,
			}
		}
	}

	SnapErr, SnapReq := FormattingTransaksi(
		ctx,
		model,
		data.AlamatInformation,
		data.DataCheckout,
		db,
		data.PaymentMethod,
		data.DataCheckout.LayananPengiriman.JenisLayananKurir,
	)

	if !SnapErr {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Message:  "Gagal server sedang sibuk coba lagi lain waktu",
		}
	}
	fmt.Println("[TRACE] FormattingTransaksi sukses, lanjut ke ValidateTransaksi()")

	SnapResponse, SnapStatus := ValidateTransaksi(SnapReq)
	if !SnapStatus {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	if SnapResponse != nil {
		fmt.Printf("[TRACE] SnapResponse Token: %s\n", SnapResponse.Token)
		fmt.Printf("[TRACE] SnapResponse RedirectURL: %s\n", SnapResponse.RedirectURL)
	} else {
		fmt.Println("[TRACE] Warning: SnapResponse nil meskipun tidak error")
	}

	fmt.Println("[TRACE] Selesai SnapTransaksi, return response ke client")
	return &response.ResponseForm{
		Status:   http.StatusOK,
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

func BatalTransaksi(ctx context.Context, data response_transaction_pengguna.SnapTransaksi, db *gorm.DB) *response.ResponseForm {
	services := "BatalTransaksi"

	var total_varian int64 = 0
	for i := 0; i < len(data.DataCheckout); i++ {
		total_varian += int64(data.DataCheckout[i].Dipesan)
	}

	var varianIds []int64 = make([]int64, 0, total_varian)
	var idkategori map[int64]int64 = make(map[int64]int64, len(data.DataCheckout))

	for i := 0; i < len(data.DataCheckout); i++ {
		idkategori[data.DataCheckout[i].IdKategoriBarang] = int64(data.DataCheckout[i].Dipesan)
		if err := db.WithContext(ctx).Model(&models.VarianBarang{}).Select("id").Where(&models.VarianBarang{
			IdBarangInduk: data.DataCheckout[i].IdBarangInduk,
			IdKategori:    data.DataCheckout[i].IdKategoriBarang,
			Status:        barang_enums.Dipesan,
			HoldBy:        data.DataCheckout[i].IDUser,
		}).Limit(int(data.DataCheckout[i].Dipesan)).Scan(&varianIds).Error; err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Message:  "Gagal coba hubungi customer service",
			}
		}
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.VarianBarang{}).
			Where("id IN ?", varianIds).
			Updates(map[string]interface{}{
				"status":        barang_enums.Ready,
				"hold_by":       0,
				"holder_entity": "",
			}).Error; err != nil {
			return err
		}

		for ind, jumlah := range idkategori {
			if err := tx.Model(&models.KategoriBarang{}).Where(&models.KategoriBarang{
				ID: ind,
			}).Update("stok", gorm.Expr("stok + ?", jumlah)).Error; err != nil {
				return err
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
			KodeTransaksiPG: pembayaran.KodeTransaksiPG,
			KodeOrderSistem: pembayaran.KodeOrderSistem,
			Provider:        pembayaran.Provider,
			Total:           pembayaran.Total,
			PaymentType:     pembayaran.PaymentType,
			PaidAt:          pembayaran.PaidAt,
		}).First(&pembayaranObj).Error; err != nil {
			fmt.Printf("[ERROR] Gagal menemukan pembayaran: %v\n", err)
			return fmt.Errorf("gagal mencari pembayaran di database: %w", err)
		}

		fmt.Printf("[TRACE] Pembayaran ditemukan: ID=%d, KodeOrder=%s\n", pembayaranObj.ID, pembayaranObj.KodeOrderSistem)

		if pembayaranObj.ID == 0 {
			fmt.Println("[ERROR] Kredensial pembayaran tidak valid (ID=0)")
			return fmt.Errorf("kredensial pembayaran tidak valid")
		}

		transaksi := models.Transaksi{
			IdPengguna:       keranjang.IDUser,
			IdSeller:         keranjang.IDSeller,
			IdBarangInduk:    int64(keranjang.IdBarangInduk),
			IdKategoriBarang: keranjang.IdKategoriBarang,
			IdAlamatPengguna: IdAlamatUser,
			IdPembayaran:     pembayaranObj.ID,
			KodeOrderSistem:  pembayaranObj.KodeOrderSistem,
			Status:           transaksi_enums.Dibayar,
			KuantitasBarang:  int32(keranjang.Dipesan),
			Total:            int64(keranjang.HargaKategori * keranjang.Dipesan),
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
		if keranjang.IDSeller == 0 || keranjang.IDUser == 0 || keranjang.IdBarangInduk == 0 {
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

		pembayaran.IdPengguna = data.DataHold[0].IDUser
		if err := tx.Create(&pembayaran).Error; err != nil {
			return err
		}

		//
		// Sanitasi Id Pengguna
		//

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

	err := db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&standard_response).Error; err != nil {
			return fmt.Errorf("gagal menyimpan PaidFailed: %w", err)
		}

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
