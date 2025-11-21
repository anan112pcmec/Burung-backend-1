package pengguna_transaction_services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"

	ekspedisi "github.com/anan112pcmec/Burung-backend-1/app/api/ekspedisi_raja_ongkir"
	ekspedisi_cost "github.com/anan112pcmec/Burung-backend-1/app/api/ekspedisi_raja_ongkir/cost"
	open_route_direction "github.com/anan112pcmec/Burung-backend-1/app/api/open_route_map/direction"
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

func GenerateItemDetail(
	ctx context.Context,
	data response_transaction_pengguna.ResponseDataCheckout,
	db *gorm.DB,
	jenisLayanan string,
	alamat_pengguna models.AlamatPengguna,
) ([]midtrans.ItemDetails, int64, []response_transaction_pengguna.DataJarak, bool) {

	var hasil []midtrans.ItemDetails
	var hargajarak []response_transaction_pengguna.DataJarak
	var total int64
	const biaya_platform = int64(5000)

	// ==== Ambil semua layanan kurir sekaligus ====
	layananRows := []models.LayananPengirimanKurir{}
	if err := db.WithContext(ctx).Find(&layananRows).Error; err != nil {
		return hasil, total, hargajarak, false
	}
	layananCache := make(map[string]int64)
	for _, r := range layananRows {
		layananCache[r.NamaLayanan] = int64(r.HargaLayanan)
	}

	idsGudang := make([]int64, len(data.DataResponse))
	for i := 0; i < len(data.DataResponse); i++ {
		idsGudang[i] = data.DataResponse[i].IdAlamatGudang
	}
	var alamatGudangList []models.AlamatGudang
	if err := db.WithContext(ctx).
		Where("id IN ? AND id_seller = ?", idsGudang, data.DataResponse[0].IDSeller).
		Take(&alamatGudangList).Error; err != nil {
		return hasil, total, hargajarak, false
	}

	alamatGudangMap := make(map[int64]models.AlamatGudang, len(data.DataResponse))
	for i := 0; i < len(alamatGudangList); i++ {
		alamatGudangMap[alamatGudangList[i].ID] = alamatGudangList[i]
	}

	// ==== Ambil biaya ongkir layanan ====
	var biayaOngkir int64
	if err := db.WithContext(ctx).Model(&models.Ongkir{}).
		Select("value").
		Where("nama = ?", strings.ToLower(jenisLayanan)).
		Limit(1).Scan(&biayaOngkir).Error; err != nil {
		biayaOngkir = 0
	}

	// ==== LOOP ITEM ====
	distanceCache := make(map[string]response_transaction_pengguna.DataJarak)
	var biayaKendaraanTotal int64
	var pengirimanCount int64

	for _, item := range data.DataResponse {
		subtotal := int64(item.HargaKategori) * int64(item.Dipesan)
		total += subtotal

		beratTotal := float64(item.BeratKategori) * float64(item.Dipesan) / 1000

		// ==== Jenis kendaraan ====
		var jenisKendaraan string
		switch {
		case beratTotal <= 10:
			jenisKendaraan = "Motor"
		case beratTotal <= 20:
			jenisKendaraan = "Mobil"
		case beratTotal <= 30:
			jenisKendaraan = "Pickup"
		default:
			jenisKendaraan = "Truk"
		}

		biayaJasa := layananCache[jenisKendaraan]
		biayaKendaraanTotal += biayaJasa
		pengirimanCount++

		// ==== ItemDetails ====
		hasil = append(hasil, midtrans.ItemDetails{
			ID:           fmt.Sprintf("%v--%v", item.IdBarangInduk, item.IdKategoriBarang),
			Price:        int64(item.HargaKategori),
			Qty:          item.Dipesan,
			Name:         fmt.Sprintf("%s - %s", item.NamaBarang, item.NamaKategori),
			MerchantName: item.NamaSeller,
			Category:     item.JenisBarang,
		})

		// ==== Alamat gudang ====
		alamatGudang := alamatGudangMap[item.IdAlamatGudang]

		statusEkspedisi := alamatGudang.Provinsi != alamat_pengguna.Provinsi

		// ==== Hitung jarak + cache ====
		key := fmt.Sprintf("%f,%f|%f,%f",
			alamatGudang.Latitude, alamatGudang.Longitude,
			alamat_pengguna.Latitude, alamat_pengguna.Longitude,
		)

		if cached, ok := distanceCache[key]; ok {
			hargajarak = append(hargajarak, cached)
			continue
		}

		Jarak, hargaJarak, status := open_route_direction.HitungJarakHargaDirection(
			[2]float64{alamatGudang.Longitude, alamatGudang.Latitude},
			[2]float64{alamat_pengguna.Longitude, alamat_pengguna.Latitude},
		)
		if !status {
			return hasil, total, hargajarak, false
		}

		var hargaEkspedisi int64
		if statusEkspedisi {
			reqPayload := ekspedisi_cost.StarterDomesticCostReq{
				Origin:      ekspedisi.JawaCities[alamatGudang.Kota],
				Destination: ekspedisi.JawaCities[alamat_pengguna.Kota],
				Weight: strconv.FormatFloat(
					float64(item.BeratKategori)*float64(item.Dipesan)/1000,
					'f', -1, 64,
				),
				Courier: "jne",
				Price:   "lowest",
			}
			res := reqPayload.DomesticCostReq(ctx)
			hargaEkspedisi = int64(res.Data[0].Cost)
		}

		dataJarak := response_transaction_pengguna.DataJarak{
			Jarak:     Jarak,
			Harga:     float64(hargaJarak + hargaEkspedisi),
			Ekspedisi: statusEkspedisi,
		}
		hargajarak = append(hargajarak, dataJarak)
		distanceCache[key] = dataJarak
	}

	var harga_kirim int64 = 0
	for i := 0; i < len(hargajarak); i++ {
		harga_kirim += int64(hargajarak[i].Harga)
	}

	// ==== Hitung total biaya kurir + tambah ke ItemDetails ====
	totalBiayaKurir := biayaOngkir*pengirimanCount + biayaKendaraanTotal + harga_kirim
	hasil = append(hasil, midtrans.ItemDetails{
		ID:           "fee-courier",
		Price:        totalBiayaKurir,
		Qty:          1,
		Name:         "Biaya Kurir",
		MerchantName: "Courier",
		Category:     "fee",
	})
	total += totalBiayaKurir

	// ==== Biaya aplikasi ====
	hasil = append(hasil, midtrans.ItemDetails{
		ID:           "fee-app",
		Price:        biaya_platform,
		Qty:          1,
		Name:         "Biaya Aplikasi",
		MerchantName: "Platform",
		Category:     "fee",
	})
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
