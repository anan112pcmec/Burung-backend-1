package helper

import (
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/midtrans/midtrans-go"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	open_route_direction "github.com/anan112pcmec/Burung-backend-1/app/open_route_map/direction"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

func DecodeJSONBody(r *http.Request, dst interface{}) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}

// GenerateRandomDigits menghasilkan angka random dengan panjang 3-5 digit
func GenerateRandomDigits() string {
	rand.Seed(time.Now().UnixNano())

	length := rand.Intn(3) + 3

	min := intPow(10, length-1)   // contoh: 100 untuk 3 digit
	max := intPow(10, length) - 1 // contoh: 999 untuk 3 digit

	return strconv.Itoa(rand.Intn(max-min+1) + min)
}

func intPow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}

func ConvertJenisBarang(jenis string) string {
	// Map internal -> DB
	mapJenis := map[string]string{
		"Pakaian&Fashion":     "Pakaian & Fashion",
		"Kosmetik&Kecantikan": "Kosmetik & Kecantikan",
		"Elektronik&Gadget":   "Elektronik & Gadget",
		"Buku&Media":          "Buku & Media",
		"Makanan&Minuman":     "Makanan & Minuman",
		"Ibu&Bayi":            "Ibu & Bayi",
		"Mainan":              "Mainan",
		"Olahraga&Outdoor":    "Olahraga & Outdoor",
		"Otomotif&Sparepart":  "Otomotif & Sparepart",
		"RumahTangga":         "Rumah Tangga",
		"AlatTulis":           "Alat Tulis",
		"Perhiasan&Aksesoris": "Perhiasan & Aksesoris",
		"ProdukDigital":       "ProdukDigital",
		"Bangunan&Perkakas":   "Bangunan & Perkakas",
		"Musik&Instrumen":     "Musik & Instrumen",
		"Film&Broadcasting":   "Film & Broadcasting",
		"SemuaBarang":         "Semua Barang",
	}

	if val, ok := mapJenis[jenis]; ok {
		return val
	}
	// fallback kalau tidak ada mapping
	return jenis
}

func ConvertJenisBarangReverse(jenis string) string {
	// Map DB -> internal
	mapReverse := map[string]string{
		"Pakaian & Fashion":     "Pakaian&Fashion",
		"Kosmetik & Kecantikan": "Kosmetik&Kecantikan",
		"Elektronik & Gadget":   "Elektronik&Gadget",
		"Buku & Media":          "Buku&Media",
		"Makanan & Minuman":     "Makanan&Minuman",
		"Ibu & Bayi":            "Ibu&Bayi",
		"Mainan":                "Mainan",
		"Olahraga & Outdoor":    "Olahraga&Outdoor",
		"Otomotif & Sparepart":  "Otomotif&Sparepart",
		"Rumah Tangga":          "RumahTangga",
		"Alat Tulis":            "AlatTulis",
		"Perhiasan & Aksesoris": "Perhiasan&Aksesoris",
		"ProdukDigital":         "ProdukDigital",
		"Bangunan & Perkakas":   "Bangunan&Perkakas",
		"Musik & Instrumen":     "Musik&Instrumen",
		"Film & Broadcasting":   "Film&Broadcasting",
		"Semua Barang":          "SemuaBarang",
	}

	if val, ok := mapReverse[jenis]; ok {
		return val
	}
	// fallback kalau tidak ada mapping
	return jenis
}

func randomString(length int, charset string) string {
	if length <= 0 || len(charset) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(length)

	lenCharset := len(charset)
	maxAcceptable := 255 - (255 % lenCharset)

	i := 0
	for i < length {
		var b [1]byte
		if _, err := crand.Read(b[:]); err != nil {
			// fallback: gunakan math/rand jika crypto/rand error
			r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
			for ; i < length; i++ {
				sb.WriteByte(charset[r.Intn(lenCharset)])
			}
			return sb.String()
		}

		if int(b[0]) >= maxAcceptable {
			// buang hasil supaya tidak bias
			continue
		}

		idx := int(b[0]) % lenCharset
		sb.WriteByte(charset[idx])
		i++
	}

	return sb.String()
}

func GenerateAutoPaymentId(db *gorm.DB) (string, error) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	alphanum := letters + numbers

	part1 := randomString(4, letters)  // 4 huruf
	part2 := randomString(5, numbers)  // 5 angka
	part3 := randomString(6, alphanum) // 6 huruf/angka

	final := part1 + "-" + part2 + "-" + part3

	var ada int64 = 0
	err_kode := db.Model(models.Transaksi{}).Where(models.Transaksi{KodeOrderSistem: final}).Count(&ada).Limit(2).Error
	if err_kode != nil {
		return final, err_kode
	}

	if ada != 0 {
		return "", err_kode
	}

	return final, nil
}

func Hitungtotal(input []int64) int64 {
	var total int64 = 0
	for _, biaya := range input {
		total += biaya
	}
	return total
}

func GenerateItemDetail(
	data response_transaction_pengguna.ResponseDataCheckout,
	db *gorm.DB,
	jenisLayanan string,
	alamat_pengguna models.AlamatPengguna,
) ([]midtrans.ItemDetails, int64) {

	var hasil []midtrans.ItemDetails
	var total int64
	var pengiriman_count int
	const biaya_platform = int64(5000)
	var biayaKendaraan []int64

	// ==== LOOP ITEM BARANG ====
	for _, item := range data.DataResponse {
		subtotal := int64(item.HargaKategori) * int64(item.Dipesan)

		var kategori models.KategoriBarang
		_ = db.Model(&models.KategoriBarang{}).
			Where(&models.KategoriBarang{ID: item.IdKategoriBarang}).
			Select("berat_gram", "dimensi_lebar_cm", "dimensi_panjang_cm", "id_alamat_gudang").
			Take(&kategori).Error

		beratTotal := float64(kategori.BeratGram) * float64(item.Dipesan) / 1000

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
		_ = db.Model(&models.LayananPengirimanKurir{}).
			Where(&models.LayananPengirimanKurir{NamaLayanan: layanan}).
			Select("harga_layanan").
			Scan(&biayaJasa)

		itemDetail := midtrans.ItemDetails{
			ID:           fmt.Sprintf("%v--%v", item.IdBarangInduk, item.IdKategoriBarang),
			Price:        int64(item.HargaKategori),
			Qty:          item.Dipesan,
			Name:         fmt.Sprintf("%s - %s", item.NamaBarang, item.NamaKategori),
			MerchantName: item.NamaSeller,
			Category:     item.JenisBarang,
		}

		hasil = append(hasil, itemDetail)
		total += subtotal
		biayaKendaraan = append(biayaKendaraan, biayaJasa)
		pengiriman_count++
	}

	// ==== AMBIL ONGKIR DASAR ====
	var biayaOngkir int64
	err := db.Model(&models.Ongkir{}).
		Where(&models.Ongkir{Nama: strings.ToLower(jenisLayanan)}).
		Select("value").
		Scan(&biayaOngkir).Error
	if err != nil {
		biayaOngkir = 0
	}

	// ==== HITUNG HARGA JARAK ====
	var HargaKirimPerJarakBarang int64

	for i, item := range data.DataResponse {
		log.Printf("[TRACE] ========== MULAI PROSES ITEM KE-%d ==========", i+1)
		log.Printf("[TRACE] ID BarangInduk: %v | ID KategoriBarang: %v", item.IdBarangInduk, item.IdKategoriBarang)

		// --- Ambil ID alamat gudang ---
		var kategori models.KategoriBarang
		err := db.Model(&models.KategoriBarang{}).
			Where(&models.KategoriBarang{ID: item.IdKategoriBarang}).
			Select("id_alamat_gudang").
			Take(&kategori).Error
		if err != nil {
			log.Printf("[TRACE] Gagal ambil KategoriBarang ID %v: %v", item.IdKategoriBarang, err)
			continue
		}
		if kategori.IDAlamat == 0 {
			log.Printf("[TRACE] ID Alamat Gudang = 0 untuk KategoriBarang ID %v, skip", item.IdKategoriBarang)
			continue
		}
		log.Printf("[TRACE] ID Alamat Gudang ditemukan: %v", kategori.IDAlamat)

		// --- Ambil data alamat gudang ---
		var alamatGudang models.AlamatGudang
		err = db.Model(&models.AlamatGudang{}).
			Where(&models.AlamatGudang{ID: kategori.IDAlamat}).
			Take(&alamatGudang).Error
		if err != nil {
			log.Printf("[TRACE] Gagal ambil AlamatGudang ID %v: %v", kategori.IDAlamat, err)
			continue
		}
		log.Printf("[TRACE] Lokasi Gudang: Lat %.6f | Lon %.6f", alamatGudang.Latitude, alamatGudang.Longitude)

		// --- Tentukan lat/long tujuan ---
		var latTujuan, longTujuan float64
		if alamat_pengguna.Latitude != 0 && alamat_pengguna.Longitude != 0 {
			latTujuan = alamat_pengguna.Latitude
			longTujuan = alamat_pengguna.Longitude
			log.Printf("[TRACE] Gunakan koordinat tujuan dari input: Lat %.6f | Lon %.6f", latTujuan, longTujuan)
		} else {
			var alamatUser models.AlamatPengguna
			err = db.Model(&models.AlamatPengguna{}).
				Where(&models.AlamatPengguna{ID: alamat_pengguna.ID}).
				Take(&alamatUser).Error
			if err != nil {
				log.Printf("[TRACE] Gagal ambil fallback AlamatPengguna ID %v: %v", alamat_pengguna.ID, err)
				continue
			}
			latTujuan = alamatUser.Latitude
			longTujuan = alamatUser.Longitude
			log.Printf("[TRACE] Gunakan koordinat fallback dari DB: Lat %.6f | Lon %.6f", latTujuan, longTujuan)
		}

		// --- Validasi koordinat sebelum request ---
		if alamatGudang.Latitude == 0 || alamatGudang.Longitude == 0 {
			log.Printf("[TRACE] Gudang belum punya koordinat lengkap (Lat/Lon = 0), skip")
			continue
		}
		if latTujuan == 0 || longTujuan == 0 {
			log.Printf("[TRACE] Tujuan belum punya koordinat lengkap (Lat/Lon = 0), skip")
			continue
		}

		// --- Hitung jarak dan harga ---
		log.Printf("[TRACE] Hitung jarak: dari (%.6f, %.6f) ke (%.6f, %.6f)",
			alamatGudang.Latitude, alamatGudang.Longitude, latTujuan, longTujuan)

		_, hargaJarak, status := open_route_direction.HitungJarakHargaDirection(
			[2]float64{alamatGudang.Longitude, alamatGudang.Latitude},
			[2]float64{longTujuan, latTujuan},
		)

		if !status {
			log.Printf("[TRACE] Hitung jarak gagal untuk item ID %v | Koordinat Gudang (%.6f, %.6f) | Tujuan (%.6f, %.6f)",
				item.IdKategoriBarang, alamatGudang.Latitude, alamatGudang.Longitude, latTujuan, longTujuan)
			continue
		}

		HargaKirimPerJarakBarang += hargaJarak
		log.Printf("[TRACE] Harga jarak item ke-%d: %v | Total sementara: %v", i+1, hargaJarak, HargaKirimPerJarakBarang)
		log.Printf("[TRACE] ========== SELESAI ITEM KE-%d ==========", i+1)
	}

	// ==== HITUNG TOTAL BIAYA ====
	totalBiayaKendaraan := Hitungtotal(biayaKendaraan)
	totalBiayaKurir := biayaOngkir*int64(pengiriman_count) + totalBiayaKendaraan - biaya_platform + HargaKirimPerJarakBarang

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

	return hasil, total
}

func UpdateSocialMediaDispatch(data models.EntitySocialMedia) []string {
	var hasil []string
	if data.Instagram != "" {
		hasil = append(hasil, "instagram")
	}

	if data.Whatsapp != "" {
		hasil = append(hasil, "whatsapp")
	}

	if data.Facebook != "" {
		hasil = append(hasil, "facebook")
	}

	if data.TikTok != "" {
		hasil = append(hasil, "tiktok")

	}

	return hasil
}

func GenerateOTP() string {
	otp := ""
	for i := 0; i < 8; i++ {
		// ambil angka random 0–9
		n, _ := crand.Int(crand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", n.Int64())
	}
	return otp
}
