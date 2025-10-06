package helper

import (
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/midtrans/midtrans-go"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
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
	err_kode := db.Model(models.Transaksi{}).Where(models.Transaksi{KodeOrder: final}).Count(&ada).Limit(2).Error
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

func GenerateItemDetail(data response_transaction_pengguna.ResponseDataCheckout, db *gorm.DB, jenisLayanan string) ([]midtrans.ItemDetails, int64) {
	var hasil []midtrans.ItemDetails
	var total int64 = 0
	var pengiriman_count = 0
	const biaya_platform = 5000
	var biayaKendaraan []int64

	for _, Item := range data.DataResponse {
		subtotal := int64(Item.HargaKategori) * int64(Item.Dipesan)

		var kategorinya models.KategoriBarang
		_ = db.Model(&models.KategoriBarang{}).
			Where(&models.KategoriBarang{ID: Item.IdKategoriBarang}).
			Select("berat_gram", "dimensi_lebar_cm", "dimensi_panjang_cm", "id_alamat_gudang").
			Take(&kategorinya).Error

		beratTotalBarangPengirian := kategorinya.BeratGram * int16(Item.Dipesan) / 1000

		var biayajk int64
		var layanan string
		switch {
		case beratTotalBarangPengirian <= 10:
			layanan = "Motor"
		case beratTotalBarangPengirian <= 20:
			layanan = "Mobil"
		case beratTotalBarangPengirian <= 30:
			layanan = "Pickup"
		default:
			layanan = "Truk"
		}

		_ = db.Model(&models.LayananPengirimanKurir{}).
			Where("nama_layanan = ?", layanan).
			Select("harga_layanan").
			Scan(&biayajk)

		itemDetail := midtrans.ItemDetails{
			ID:           fmt.Sprintf("%v--%v", Item.IdBarangInduk, Item.IdKategoriBarang),
			Price:        int64(Item.HargaKategori),
			Qty:          Item.Dipesan,
			Name:         fmt.Sprintf("%s - %s", Item.NamaBarang, Item.NamaKategori),
			MerchantName: Item.NamaSeller,
			Category:     Item.JenisBarang,
		}

		pengiriman_count++
		hasil = append(hasil, itemDetail)
		total += subtotal
		biayaKendaraan = append(biayaKendaraan, biayajk)
	}

	var biaya_ongkir int64
	err := db.Model(&models.Ongkir{}).
		Select("value").
		Where("nama = ?", strings.ToLower(jenisLayanan)).
		Scan(&biaya_ongkir).Error
	if err != nil {
		biaya_ongkir = 0
	}

	totalBiayaKendaraan := Hitungtotal(biayaKendaraan)
	totalBiayaKurir := biaya_ongkir*int64(pengiriman_count) + totalBiayaKendaraan - biaya_platform

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
