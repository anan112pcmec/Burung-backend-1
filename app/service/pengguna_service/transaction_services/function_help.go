package pengguna_transaction_services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	open_route_direction "github.com/anan112pcmec/Burung-backend-1/app/open_route_map/direction"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

func GenerateItemDetail(
	ctx context.Context,
	data response_transaction_pengguna.ResponseDataCheckout,
	db *gorm.DB,
	jenisLayanan string,
	alamat_pengguna models.AlamatPengguna,
) ([]midtrans.ItemDetails, int64, []response_transaction_pengguna.DataJarak, bool) {

	var hasil []midtrans.ItemDetails = make([]midtrans.ItemDetails, 0, len(data.DataResponse))
	var total int64
	var pengiriman_count int
	const biaya_platform = int64(5000)
	var biayaKendaraan []int64 = make([]int64, 0, len(data.DataResponse))
	var hargajarak []response_transaction_pengguna.DataJarak = make([]response_transaction_pengguna.DataJarak, 0, len(data.DataResponse))

	for i := 0; i < len(data.DataResponse); i++ {
		subtotal := int64(data.DataResponse[i].HargaKategori) * int64(data.DataResponse[i].Dipesan)

		beratTotal := float64(data.DataResponse[i].BeratKategori) * float64(data.DataResponse[i].Dipesan) / 1000

		var layanan string
		switch {
		case beratTotal <= 10:
			layanan = "Motor"
		case beratTotal <= 20:
			layanan = "Mobil"
		case beratTotal <= 30:
			layanan = "Pickup"
		default:
			layanan = "Truk"
		}

		var biayaJasa int64
		if err := db.WithContext(ctx).Model(&models.LayananPengirimanKurir{}).
			Where(&models.LayananPengirimanKurir{NamaLayanan: layanan}).
			Select("harga_layanan").Limit(1).
			Scan(&biayaJasa).Error; err != nil {
			return hasil, total, hargajarak, false
		}

		itemDetail := midtrans.ItemDetails{
			ID:           fmt.Sprintf("%v--%v", data.DataResponse[i].IdBarangInduk, data.DataResponse[i].IdKategoriBarang),
			Price:        int64(data.DataResponse[i].HargaKategori),
			Qty:          data.DataResponse[i].Dipesan,
			Name:         fmt.Sprintf("%s - %s", data.DataResponse[i].NamaBarang, data.DataResponse[i].NamaKategori),
			MerchantName: data.DataResponse[i].NamaSeller,
			Category:     data.DataResponse[i].JenisBarang,
		}

		hasil = append(hasil, itemDetail)
		total += subtotal
		biayaKendaraan = append(biayaKendaraan, biayaJasa)
		pengiriman_count++
	}

	var biayaOngkir int64
	err := db.WithContext(ctx).Model(&models.Ongkir{}).Select("value").
		Where(&models.Ongkir{Nama: strings.ToLower(jenisLayanan)}).
		Limit(1).
		Scan(&biayaOngkir).Error
	if err != nil {
		biayaOngkir = 0
	}

	// ==== HITUNG HARGA JARAK ====
	var HargaKirimPerJarakBarang int64
	distanceCache := make(map[string]int64) // 🚀 CACHE JARAK

	for i := 0; i < len(data.DataResponse); i++ {
		log.Printf("[TRACE] ========== MULAI PROSES ITEM KE-%d ==========", i+1)
		log.Printf("[TRACE] ID BarangInduk: %v | ID KategoriBarang: %v", data.DataResponse[i].IdBarangInduk, data.DataResponse[i].IdKategoriBarang)

		// --- Ambil data alamat gudang ---
		var alamatGudang models.AlamatGudang
		err = db.Model(&models.AlamatGudang{}).
			Where(&models.AlamatGudang{ID: data.DataResponse[i].IdAlamatGudang}).
			Take(&alamatGudang).Error
		if err != nil {
			continue
		}

		var latTujuan, longTujuan float64
		if alamat_pengguna.Latitude != 0 && alamat_pengguna.Longitude != 0 {
			latTujuan = alamat_pengguna.Latitude
			longTujuan = alamat_pengguna.Longitude
		} else {
			var alamatUser models.AlamatPengguna
			err = db.WithContext(ctx).Model(&models.AlamatPengguna{}).
				Where(&models.AlamatPengguna{ID: alamat_pengguna.ID}).Limit(1).
				Take(&alamatUser).Error
			if err != nil {
				return hasil, total, hargajarak, false
			}
			latTujuan = alamatUser.Latitude
			longTujuan = alamatUser.Longitude
		}

		if alamatGudang.Latitude == 0 || alamatGudang.Longitude == 0 {
			return hasil, total, hargajarak, false
		}
		if latTujuan == 0 || longTujuan == 0 {
			return hasil, total, hargajarak, false
		}

		// 🚀 =============== OPTIMASI DISINI =================
		key := fmt.Sprintf("%f,%f|%f,%f",
			alamatGudang.Latitude, alamatGudang.Longitude,
			latTujuan, longTujuan)

		if cached, ok := distanceCache[key]; ok {
			hargajarak = append(hargajarak, hargajarak[i-1])

			HargaKirimPerJarakBarang += cached
			log.Printf("[TRACE] ❗ CACHE digunakan untuk item ke-%d = %v", i+1, cached)
			continue
		}
		// ====================================================

		Jarak, hargaJarak, status := open_route_direction.HitungJarakHargaDirection(
			[2]float64{alamatGudang.Longitude, alamatGudang.Latitude},
			[2]float64{longTujuan, latTujuan},
		)

		hargajarak = append(hargajarak, response_transaction_pengguna.DataJarak{
			Jarak: Jarak,
			Harga: float64(hargaJarak),
		})

		if !status {
			return hasil, total, hargajarak, false
		}

		fmt.Println("Ini harga jarak cuy:", hargaJarak)

		// simpan ke cache
		distanceCache[key] = hargaJarak
		HargaKirimPerJarakBarang += hargaJarak

		log.Printf("[TRACE] Harga jarak item ke-%d: %v | Total sementara: %v",
			i+1, hargaJarak, HargaKirimPerJarakBarang)
	}

	// ==== HITUNG TOTAL BIAYA ====
	totalBiayaKendaraan := helper.Hitungtotal(biayaKendaraan)
	totalBiayaKurir := biayaOngkir*int64(pengiriman_count) + totalBiayaKendaraan + HargaKirimPerJarakBarang

	// ==== TAMBAH ITEM BIAYA KURIR ====
	courierFee := midtrans.ItemDetails{
		ID:           "fee-courier",
		Price:        totalBiayaKurir,
		Qty:          1,
		Name:         "Biaya Kurir",
		MerchantName: "Courier",
		Category:     "fee",
	}
	hasil = append(hasil, courierFee)
	total += totalBiayaKurir

	// ==== TAMBAH ITEM BIAYA APLIKASI ====
	appFee := midtrans.ItemDetails{
		ID:           "fee-app",
		Price:        biaya_platform,
		Qty:          1,
		Name:         "Biaya Aplikasi",
		MerchantName: "Platform",
		Category:     "fee",
	}
	hasil = append(hasil, appFee)
	total += biaya_platform

	return hasil, total, hargajarak, true
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Formatting Transaksi
// Befungsi Untuk membantu snap trnsaksi dalam memformatkan sebuah transaksi
// ////////////////////////////////////////////////////////////////////////////////////

func FormattingTransaksi(ctx context.Context, user models.Pengguna, alamat models.AlamatPengguna, data response_transaction_pengguna.ResponseDataCheckout, db *gorm.DB, PaymentMethod, jenis_layanan string) (bool, *snap.Request, []response_transaction_pengguna.DataJarak) {
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

	AlamatPengguna := midtrans.CustomerAddress{
		Address:     alamat.NamaAlamat,
		City:        alamat.Kota,
		Postcode:    alamat.KodePos,
		CountryCode: alamat.KodeNegara,
	}

	fmt.Println("[TRACE] Generate ItemDetail dan TotalHarga")
	items, TotalHarga, dataJarak, status := GenerateItemDetail(ctx, data, db, jenis_layanan, alamat)
	if !status {
		return false, nil, dataJarak
	}

	if err_payment != nil {
		fmt.Println("[TRACE] Error fatal: gagal generate PaymentCode setelah retry")
		return false, nil, dataJarak
	}

	var PM []snap.SnapPaymentType
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
	return true, SnapReqeust, dataJarak
}

// ////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Validating Transaksi
// Befungsi Untuk membantu snap transaksi untuk membuat sebuah transaksi
// ////////////////////////////////////////////////////////////////////////////////////

func ValidateTransaksi(snapReq *snap.Request) (*snap.Response, bool) {
	fmt.Println("[TRACE] Start ValidateTransaksi")

	var s snap.Client
	s.New("Mid-server-7wpABbBW_WURdLxcxc5bX5eb", midtrans.Sandbox)

	fmt.Println("[TRACE] Membuat transaksi dengan Snap SDK")
	snapResp, err := s.CreateTransaction(snapReq)
	if err != nil {
		fmt.Printf("[TRACE] Gagal membuat transaksi: %v\n", err)
		return nil, false
	}

	fmt.Println("[TRACE] Selesai ValidateTransaksi, return response")
	return snapResp, true
}
