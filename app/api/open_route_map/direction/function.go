package open_route_direction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

func Getenvi(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var ApiKey = Getenvi("eyJvcmciOiI1YjNjZTM1OTc4NTExMTAwMDFjZjYyNDgiLCJpZCI6ImMxNTVhOTVhZmQ0NDQ4M2M5OTg4ODEyYmZkMjY4ODUyIiwiaCI6Im11cm11cjY0In0=", "eyJvcmciOiI1YjNjZTM1OTc4NTExMTAwMDFjZjYyNDgiLCJpZCI6ImMxNTVhOTVhZmQ0NDQ4M2M5OTg4ODEyYmZkMjY4ODUyIiwiaCI6Im11cm11cjY0In0=")

func HitungJarakHargaDirection(TitikMulai [2]float64, TitikSelesai [2]float64) (float64, int64, bool) {
	var (
		Jarak  float64 = 0
		Harga  int64   = 0
		Status bool    = false
	)

	TitikMulai[0] = math.Round(TitikMulai[0]*1e4) / 1e4
	TitikMulai[1] = math.Round(TitikMulai[1]*1e4) / 1e4
	TitikSelesai[0] = math.Round(TitikSelesai[0]*1e4) / 1e4
	TitikSelesai[1] = math.Round(TitikSelesai[1]*1e4) / 1e4

	fmt.Println("=== [TRACE] Mulai HitungJarakHargaDirection ===")
	fmt.Printf("[INPUT] TitikMulai  : %.6f, %.6f\n", TitikMulai[0], TitikMulai[1])
	fmt.Printf("[INPUT] TitikSelesai: %.6f, %.6f\n", TitikSelesai[0], TitikSelesai[1])

	// --- Siapkan request body ---
	reqBody := RequestDirection{
		Koordinat: [][2]float64{
			TitikMulai, TitikSelesai,
		},
		Bahasa:       "id",
		Satuan:       "m",
		Geometry:     false,
		Instructions: false,
	}

	fmt.Println("[TRACE] Membuat body request JSON...")
	body, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		fmt.Println("[ERROR] Gagal encode JSON:", err)
		return Jarak, Harga, false
	}
	fmt.Println("[TRACE] JSON body siap dikirim:")
	fmt.Println(string(body))

	// --- Kirim request ke ORS ---
	url := "https://api.openrouteservice.org/v2/directions/driving-car"
	fmt.Println("[TRACE] Mengirim request POST ke:", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("[ERROR] Gagal buat request:", err)
		return Jarak, Harga, false
	}

	req.Header.Set("Authorization", ApiKey)
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("[TRACE] Header Authorization dan Content-Type diset.")

	client := &http.Client{}
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[ERROR] Gagal kirim request:", err)
		return Jarak, Harga, false
	}
	defer resp.Body.Close()
	fmt.Printf("[TRACE] Request selesai dalam %.2f detik\n", time.Since(startTime).Seconds())

	data, _ := io.ReadAll(resp.Body)
	fmt.Println("[TRACE] Status kode dari API:", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("[ERROR] Respon dari API (bukan 200):")
		fmt.Println(string(data))
		return Jarak, Harga, false
	}

	// --- Decode hasil JSON ---
	fmt.Println("[TRACE] Mendecode hasil respon JSON...")
	var hasil ResponseDirection
	if err := json.Unmarshal(data, &hasil); err != nil {
		fmt.Println("[ERROR] Gagal decode JSON:", err)
		return Jarak, Harga, false
	}

	// --- Ambil jarak total ---
	fmt.Println(hasil)
	if len(hasil.Routes) > 0 {
		distanceMeter := hasil.Routes[0].Summary.Distance
		durationSecond := hasil.Routes[0].Summary.Duration
		jarakKm := distanceMeter / 1000
		durasiMenit := durationSecond / 60

		Jarak = jarakKm
		Harga = int64(jarakKm * 1000) // Rp 1000/km
		Status = true

		fmt.Println("=== [TRACE HASIL] ===")
		fmt.Printf("Jarak total   : %.2f km\n", jarakKm)
		fmt.Printf("Durasi waktu  : %.2f menit\n", durasiMenit)
		fmt.Printf("Perkiraan harga: Rp %d\n", Harga)
		fmt.Println("=====================")
	} else {
		fmt.Println("[WARN] Tidak ada data features di response.")
	}

	fmt.Println("=== [TRACE] Selesai HitungJarakHargaDirection ===")
	return Jarak, Harga, Status
}
