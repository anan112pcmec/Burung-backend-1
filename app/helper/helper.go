package helper

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func DecodeJSONBody(r *http.Request, dst interface{}) error {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // kalau ada field asing diabaikan

	// coba decode, tapi jangan maksa kalau ada field kosong
	if err := dec.Decode(dst); err != nil {
		// kalau body kosong total
		if errors.Is(err, io.EOF) {
			return nil // biarin struct dst default
		}
		return err // JSON rusak atau tipe salah
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
